package main

import (
	"encoding/json"
	"log"
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

	AngAccuracy Angle // Accuracy of the measurement in radians.

	LongSideFOV     Angle // The field of view of the longest side of every image in radians.
	LongSideFOVLock bool  // Prevent the value from being optimized.

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
		LongSideFOV: 1.2, // Start with ~70 deg of FOV.
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

	if !c.LongSideFOVLock {
		tweakables = append(tweakables, &c.LongSideFOV)
	}

	for _, photo := range c.Photos {
		newTweakables, newResiduals := photo.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}
	return tweakables, residuals
}

func (c *Camera) GetProjectionMatrix(width, height float64) mgl64.Mat4 {
	aspect := width / height

	var fovy float64
	if width > height {
		fovy = float64(c.LongSideFOV) / aspect // BUG: FOVY calculation is wrong, it has to use the atan somehow
	} else {
		fovy = float64(c.LongSideFOV)
	}

	return mgl64.Perspective(fovy, aspect, 0.001, 1)
}
