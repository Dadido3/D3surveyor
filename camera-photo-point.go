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

import "time"

// CameraPhotoPoint maps a point onto a photo.
type CameraPhotoPoint struct {
	photo *CameraPhoto
	key   string

	CreatedAt time.Time

	PointKey string // The unique ID of the point.

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

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (p *CameraPhotoPoint) Copy() *CameraPhotoPoint {
	return &CameraPhotoPoint{
		CreatedAt:  p.CreatedAt,
		PointKey:   p.PointKey,
		X:          p.X,
		Y:          p.Y,
		projectedX: p.projectedX,
		projectedY: p.projectedY,
		sr:         p.sr,
		Suggested:  p.Suggested,
	}
}
