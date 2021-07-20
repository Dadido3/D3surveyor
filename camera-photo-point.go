package main

import "time"

// CameraPhotoPoint maps a point onto a photo.
type CameraPhotoPoint struct {
	photo *CameraPhoto
	key   string

	CreatedAt time.Time

	point string // The unique ID of the point.

	x, y float64 // Point's position on the photo in the range of [0,1]
}

func (cp *CameraPhoto) NewPoint() *CameraPhotoPoint {
	key := cp.camera.site.shortIDGen.MustGenerate()

	point := &CameraPhotoPoint{
		photo:     cp,
		key:       key,
		CreatedAt: time.Now(),
	}

	cp.points[key] = point

	return point
}

func (p *CameraPhotoPoint) Key() string {
	return p.key
}

func (p *CameraPhotoPoint) Delete() {
	delete(p.photo.points, p.Key())
}
