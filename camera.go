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

	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
	"github.com/vugu/vugu/js"
)

// Amount of camera distortion coefficients.
const CameraDistortionKs = 2

type Camera struct {
	vgrouter.NavigatorRef `json:"-"`

	site *Site
	key  string

	Name      string
	CreatedAt time.Time

	PixelAccuracy PixelDistance // Accuracy of the measurement.

	HorizontalAOV       Angle // The horizontal angle of view of the camera.
	HorizontalAOVLocked bool  // Prevent the value from being optimized.

	// Lens distortion model parameters.

	DistortionCenterOffset       PixelCoordinate                    // Image center offset measured from the real image center. (Offset of the principal point)
	DistortionCenterOffsetLocked bool                               // Locked state of the image center offset.
	DistortionKs                 [CameraDistortionKs]TweakableFloat // List of distortion coefficients.
	DistortionKsLocked           [CameraDistortionKs]bool           // Locked state of the distortion coefficients.

	Photos map[string]*CameraPhoto
}

func (s *Site) NewCamera(name string) *Camera {
	key := s.shortIDGen.MustGenerate()

	c := &Camera{
		site:                         s,
		key:                          key,
		Name:                         name,
		CreatedAt:                    time.Now(),
		PixelAccuracy:                100,
		HorizontalAOV:                70 * 2 * math.Pi / 360, // Start with a guess of 70 deg for AOV.
		DistortionCenterOffsetLocked: true,
		DistortionKsLocked:           [2]bool{true, true},
		Photos:                       map[string]*CameraPhoto{},
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

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (c *Camera) Copy() *Camera {
	copy := &Camera{
		Name:                         c.Name,
		CreatedAt:                    c.CreatedAt,
		PixelAccuracy:                c.PixelAccuracy,
		HorizontalAOV:                c.HorizontalAOV,
		HorizontalAOVLocked:          c.HorizontalAOVLocked,
		DistortionCenterOffset:       c.DistortionCenterOffset,
		DistortionCenterOffsetLocked: c.DistortionCenterOffsetLocked,
		DistortionKs:                 c.DistortionKs,
		DistortionKsLocked:           c.DistortionKsLocked,
		Photos:                       map[string]*CameraPhoto{},
	}

	// Generate copies of all children.
	for k, v := range c.Photos {
		copy.Photos[k] = v.Copy()
	}

	// Restore keys and references.
	copy.RestoreChildrenRefs()

	return copy
}

// RestoreChildrenRefs updates the key of the children and any reference to this object.
func (c *Camera) RestoreChildrenRefs() {
	for k, v := range c.Photos {
		v.key, v.camera = k, c
	}
}

func (c *Camera) UnmarshalJSON(data []byte) error {
	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Camera
	if err := json.Unmarshal(data, tempType(c)); err != nil {
		return err
	}

	// Restore keys and references.
	c.RestoreChildrenRefs()

	return nil
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (c *Camera) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}

	if !c.HorizontalAOVLocked {
		tweakables = append(tweakables, &c.HorizontalAOV)
	}

	if !c.DistortionCenterOffsetLocked {
		tweakables = append(tweakables, &c.DistortionCenterOffset[0])
		tweakables = append(tweakables, &c.DistortionCenterOffset[1])
	}

	for i, locked := range c.DistortionKsLocked {
		if !locked {
			tweakables = append(tweakables, &c.DistortionKs[i])
		}
	}

	for _, photo := range c.PhotosSorted() {
		newTweakables, newResiduals := photo.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}
	return tweakables, residuals
}

// radialDistortFactor returns a radial scaling factor for the given undistorted (quared) radius.
// Any undistorted coordinate scaled with this factor around the distortion center will result in an distorted coordinate.
func (c *Camera) radialDistortFactor(radiusSqr float64) float64 {
	k1, k2 := float64(c.DistortionKs[0]), float64(c.DistortionKs[1])

	return 1 + k1*radiusSqr + k2*radiusSqr*radiusSqr
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
