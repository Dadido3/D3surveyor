// Copyright (C) 2021-2025 David Vogel
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
	js "github.com/vugu/vugu/js"
)

// Amount of camera distortion coefficients.
const CameraDistortionKs, CameraDistortionPs, CameraDistortionBs = 4, 4, 2

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
	// We will the Brown-Conrady model with the transformation direction from undistorted to distorted.
	// This is similar to what OpenCV uses, see: https://docs.opencv.org/3.4/d9/d0c/group__calib3d.html

	PrincipalPointOffset       PixelCoordinate
	PrincipalPointOffsetLocked bool
	DistortionKs               [CameraDistortionKs]TweakableFloat // List of radial distortion coefficients.
	DistortionKsLocked         [CameraDistortionKs]bool           // Locked state of the distortion coefficients.
	DistortionPs               [CameraDistortionPs]TweakableFloat // List of tangential distortion coefficients.
	DistortionPsLocked         [CameraDistortionPs]bool           // Locked state of the distortion coefficients.
	DistortionBs               [CameraDistortionBs]PixelDistance  // List of affinity and non-orthogonality distortion coefficients.
	DistortionBsLocked         [CameraDistortionBs]bool           // Locked state of the distortion coefficients.

	Photos map[string]*CameraPhoto
}

func (s *Site) NewCamera(name string) *Camera {
	c := new(Camera)
	c.initData()
	c.initReferences(s, s.shortIDGen.MustGenerate())
	c.Name = name

	return c
}

// initData initializes the object with default values and other stuff.
func (c *Camera) initData() {
	c.CreatedAt = time.Now()
	c.PixelAccuracy = 100
	c.HorizontalAOV = 70 * 2 * math.Pi / 360 // Start with a guess of 70 deg for AOV.
	c.HorizontalAOVLocked = true             // Lock AOV by default.
	c.PrincipalPointOffsetLocked = true
	c.DistortionKsLocked = [4]bool{true, true, true, true}
	c.DistortionPsLocked = [4]bool{true, true, true, true}
	c.DistortionBsLocked = [2]bool{true, true}
	c.Photos = map[string]*CameraPhoto{}
}

// initReferences updates references from and to this object and its key.
// This is only used internally to update references for copies or marshalled objects.
// This can't be used on its own to transfer an object from one parent to another.
func (c *Camera) initReferences(newParent *Site, newKey string) {
	c.site, c.key = newParent, newKey
	c.site.Cameras[c.Key()] = c
}

func (c *Camera) Key() string {
	return c.key
}

// DisplayName returns either the name, or if that is empty the key.
func (c *Camera) DisplayName() string {
	if c.Name != "" {
		return c.Name
	}

	return "(" + c.Key() + ")"
}

func (c *Camera) Delete() {
	delete(c.site.Cameras, c.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (c *Camera) Copy(newParent *Site, newKey string) *Camera {
	copy := new(Camera)
	copy.initData()
	copy.initReferences(newParent, newKey)
	copy.Name = c.Name
	copy.CreatedAt = c.CreatedAt
	copy.PixelAccuracy = c.PixelAccuracy
	copy.HorizontalAOV = c.HorizontalAOV
	copy.HorizontalAOVLocked = c.HorizontalAOVLocked
	copy.PrincipalPointOffset = c.PrincipalPointOffset
	copy.PrincipalPointOffsetLocked = c.PrincipalPointOffsetLocked
	copy.DistortionKs = c.DistortionKs
	copy.DistortionKsLocked = c.DistortionKsLocked
	copy.DistortionPs = c.DistortionPs
	copy.DistortionPsLocked = c.DistortionPsLocked
	copy.DistortionBs = c.DistortionBs
	copy.DistortionBsLocked = c.DistortionBsLocked

	// Generate copies of all children.
	for k, v := range c.Photos {
		v.Copy(copy, k)
	}

	return copy
}

func (c *Camera) UnmarshalJSON(data []byte) error {
	c.initData()

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Camera
	if err := json.Unmarshal(data, tempType(c)); err != nil {
		return err
	}

	// Update parent references and keys.
	for k, v := range c.Photos {
		v.initReferences(c, k)
	}

	return nil
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

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (c *Camera) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}

	if !c.HorizontalAOVLocked {
		tweakables = append(tweakables, &c.HorizontalAOV)
	}

	if !c.PrincipalPointOffsetLocked {
		tweakables = append(tweakables, &c.PrincipalPointOffset[0])
		tweakables = append(tweakables, &c.PrincipalPointOffset[1])
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
