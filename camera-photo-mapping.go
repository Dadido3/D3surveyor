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
	"time"
)

// CameraPhotoMapping maps a point onto a photo.
type CameraPhotoMapping struct {
	photo *CameraPhoto
	key   string

	CreatedAt time.Time

	PointKey string // The unique ID of the point.

	XMigrate float64 `json:"X"` // Value to be migrated. // TODO: Remove value migration in the future.
	YMigrate float64 `json:"Y"` // Value to be migrated. // TODO: Remove value migration in the future.

	Position     PixelCoordinate // Image position that maps the point to the photo. This is where the point should be on the image.
	projectedPos PixelCoordinate // The point's projected position. This is where the point actually is on the image.
	sr           float64         // Current squared residue value.

	Suggested bool // This mapping is just suggested, it wasn't placed or confirmed by the user (yet).
}

func (cp *CameraPhoto) NewMapping() *CameraPhotoMapping {
	key := cp.camera.site.shortIDGen.MustGenerate()

	mapping := &CameraPhotoMapping{
		photo:     cp,
		key:       key,
		CreatedAt: time.Now(),
	}

	cp.Mappings[key] = mapping

	return mapping
}

func (p *CameraPhotoMapping) Key() string {
	return p.key
}

func (p *CameraPhotoMapping) Delete() {
	delete(p.photo.Mappings, p.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (p *CameraPhotoMapping) Copy() *CameraPhotoMapping {
	return &CameraPhotoMapping{
		CreatedAt:    p.CreatedAt,
		PointKey:     p.PointKey,
		Position:     p.Position,
		projectedPos: p.projectedPos,
		sr:           p.sr,
		Suggested:    p.Suggested,
	}
}
