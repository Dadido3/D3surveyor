package main

import "time"

// CameraPhotoPoint maps a point onto a photo.
type CameraPhotoPoint struct {
	photo *CameraPhoto
	key   string

	CreatedAt time.Time

	Point string // The unique ID of the point.

	X, Y                   float64 // Point's position on the photo in the range of [0,1]. Origin is at the top left.
	projectedX, projectedY float64 // Correct position of the point on the photo in the range of [0,1]. Origin is at the top left.
	sr                     float64 // Current squared residue value.

	Suggested bool // This point is just a suggested point, not one placed or confirmed by the user (yet).
}

func (cp *CameraPhoto) NewPoint() *CameraPhotoPoint {
	key := cp.camera.site.shortIDGen.MustGenerate()

	point := &CameraPhotoPoint{
		photo:     cp,
		key:       key,
		CreatedAt: time.Now(),
	}

	cp.Points[key] = point

	return point
}

func (p *CameraPhotoPoint) Key() string {
	return p.key
}

func (p *CameraPhotoPoint) Delete() {
	delete(p.photo.Points, p.Key())
}
