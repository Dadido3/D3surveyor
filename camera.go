// Copyright (C) 2021 David Vogel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"log"
	"math"
	"sort"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
	"github.com/vugu/vugu/js"
)

type Camera struct {
	vgrouter.NavigatorRef `json:"-"`

	site *Site
	key  string

	Name      string
	CreatedAt time.Time

	AngAccuracy Angle // Accuracy of the measurement.

	LongSideAOV     Angle // The angle of view of the longest side of every image.
	LongSideAOVLock bool  // Prevent the value from being optimized.

	Photos map[string]*CameraPhoto
}

func (s *Site) NewCamera(name string) *Camera {
	key := s.shortIDGen.MustGenerate()

	c := &Camera{
		site:        s,
		key:         key,
		Name:        name,
		CreatedAt:   time.Now(),
		AngAccuracy: 0.2, // Assume ~10 deg of accuracy.
		LongSideAOV: 1.2, // Start with ~70 deg of AOV.
		Photos:      map[string]*CameraPhoto{},
	}

	s.Cameras[key] = c

	return c
}

func (c *Camera) handleFileChange(event vugu.DOMEvent) {
	fileReader := js.Global().Get("FileReader").New()
	fileReader.Call("addEventListener", "loadend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		buffer := fileReader.Get("result")
		uint8Array := js.Global().Get("Uint8Array").New(buffer)

		imageData := make([]byte, uint8Array.Length())
		js.CopyBytesToGo(imageData, uint8Array)

		event.EventEnv().Lock()
		defer event.EventEnv().UnlockRender()

		photo, err := c.NewPhoto(imageData)
		if err != nil {
			log.Printf("Couldn't load image: %v", err)
			// TODO: Somehow tell the user the image couldn't be loaded
			return nil
		}

		c.Navigate("/camera/"+c.Key()+"/photo/"+photo.Key(), nil)

		return js.Undefined()
	}))

	imgFile := js.Global().Get("document").Call("getElementById", "photo-upload").Get("files").Index(0)
	fileReader.Call("readAsArrayBuffer", imgFile)
}

func (c *Camera) Key() string {
	return c.key
}

func (c *Camera) Delete() {
	delete(c.site.Cameras, c.Key())
}

func (c *Camera) UnmarshalJSON(data []byte) error {
	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Camera
	if err := json.Unmarshal(data, tempType(c)); err != nil {
		return err
	}

	// Restore keys and references.
	for k, v := range c.Photos {
		v.key, v.camera = k, c
	}

	return nil
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (c *Camera) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}

	if !c.LongSideAOVLock {
		tweakables = append(tweakables, &c.LongSideAOV)
	}

	for _, photo := range c.Photos {
		newTweakables, newResiduals := photo.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}
	return tweakables, residuals
}

func (c *Camera) GetProjectionMatrix(width, height float64) mgl64.Mat4 {
	aspect := width / height

	var aovY float64
	if width > height {
		aovY = 2 * math.Atan(math.Tan(float64(c.LongSideAOV)*0.5)/aspect)
	} else {
		aovY = float64(c.LongSideAOV)
	}

	return mgl64.Perspective(aovY, aspect, 0.001, 1)
}

// PhotosSorted returns the photos of the camera as a list sorted by date.
// TODO: Replace with generics once they are available. It's one of the few cases where they are really needed
func (s *Camera) PhotosSorted() []*CameraPhoto {
	photos := make([]*CameraPhoto, 0, len(s.Photos))

	for _, photo := range s.Photos {
		photos = append(photos, photo)
	}

	sort.Slice(photos, func(i, j int) bool {
		return photos[i].CreatedAt.After(photos[j].CreatedAt)
	})

	return photos
}
