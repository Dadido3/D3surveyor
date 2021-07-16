package main

import (
	"bytes"
	"image"
	"time"

	"github.com/vugu/vgrouter"
)

type CameraPhoto struct {
	vgrouter.NavigatorRef
	camera *Camera
	key    string

	CreatedAt time.Time

	ImageData []byte
	ImageConf image.Config
}

func (c *Camera) NewPhoto(imageData []byte) (*CameraPhoto, error) {
	imageConf, _, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	key := c.site.shortIDGen.MustGenerate()

	cp := &CameraPhoto{
		camera:    c,
		key:       key,
		CreatedAt: time.Now(),
		ImageData: imageData,
		ImageConf: imageConf,
	}

	c.Photos[key] = cp

	return cp, nil
}

func (cp *CameraPhoto) Key() string {
	return cp.key
}

func (cp *CameraPhoto) Delete() {
	delete(cp.camera.Photos, cp.Key())
}
