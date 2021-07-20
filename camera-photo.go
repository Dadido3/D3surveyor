package main

import (
	"bytes"
	"image"
	"time"

	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu/js"
)

type CameraPhoto struct {
	vgrouter.NavigatorRef
	camera *Camera
	key    string

	CreatedAt time.Time

	ImageData []byte // TODO: Don't store the image as byte slice. Only store it as a js blob.
	ImageConf image.Config

	jsImageBlob js.Value // Blob representing the image on the js side.
	jsImageURL  js.Value // URL referencing the blob.

	points map[string]*CameraPhotoPoint // List of points mapped to this photo.
}

func (c *Camera) NewPhoto(imageData []byte) (*CameraPhoto, error) {
	imageConf, _, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	key := c.site.shortIDGen.MustGenerate()

	// Create js blob and URL.
	dst := js.Global().Get("Uint8Array").New(len(imageData))
	js.CopyBytesToJS(dst, imageData)
	dstArray := js.Global().Get("Array").New(dst)
	blob := js.Global().Get("Blob").New(dstArray, js.ValueOf(map[string]interface{}{"type": "image/*"}))
	url := js.Global().Get("URL").Call("createObjectURL", blob) // This has to be freed when the photo is deleted.

	cp := &CameraPhoto{
		camera:      c,
		key:         key,
		CreatedAt:   time.Now(),
		ImageData:   imageData,
		ImageConf:   imageConf,
		jsImageBlob: blob,
		jsImageURL:  url,
		points:      map[string]*CameraPhotoPoint{},
	}

	c.Photos[key] = cp

	return cp, nil
}

func (cp *CameraPhoto) Key() string {
	return cp.key
}

func (cp *CameraPhoto) Delete() {
	delete(cp.camera.Photos, cp.Key())

	js.Global().Get("URL").Call("revokeObjectURL", cp.jsImageURL)
}

func (cp *CameraPhoto) JsImageURL() string {
	return cp.jsImageURL.String()
}
