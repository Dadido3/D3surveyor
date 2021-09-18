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
	"time"
)

// CameraPhotoMapping maps a point onto a photo.
type CameraPhotoMapping struct {
	photo *CameraPhoto
	key   string

	CreatedAt time.Time

	PointKey string // The unique ID of the point.

	Position     PixelCoordinate // Image position that maps the point to the photo. This is where the point should be projected.
	projectedPos PixelCoordinate // The point's projected position. This is where the point actually is projected.
	sr           float64         // Current squared residual value.

	Suggested bool // This mapping is just suggested, it wasn't placed or confirmed by the user (yet).
}

func (cp *CameraPhoto) NewMapping() *CameraPhotoMapping {
	m := new(CameraPhotoMapping)
	m.initData()
	m.initReferences(cp, cp.camera.site.shortIDGen.MustGenerate())

	return m
}

// initData initializes the object with default values and other stuff.
func (m *CameraPhotoMapping) initData() {
	m.CreatedAt = time.Now()
}

// initReferences updates references from and to this object and its key.
// This is only used internally to update references for copies or marshalled objects.
// This can't be used on its own to transfer an object from one parent to another.
func (m *CameraPhotoMapping) initReferences(newParent *CameraPhoto, newKey string) {
	m.photo, m.key = newParent, newKey
	m.photo.Mappings[m.Key()] = m
}

func (m *CameraPhotoMapping) Key() string {
	return m.key
}

func (m *CameraPhotoMapping) Delete() {
	delete(m.photo.Mappings, m.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (m *CameraPhotoMapping) Copy(newParent *CameraPhoto, newKey string) *CameraPhotoMapping {
	copy := new(CameraPhotoMapping)
	copy.initData()
	copy.initReferences(newParent, newKey)
	copy.CreatedAt = m.CreatedAt
	copy.PointKey = m.PointKey
	copy.Position = m.Position
	copy.projectedPos = m.projectedPos
	copy.sr = m.sr
	copy.Suggested = m.Suggested

	return copy
}

func (m *CameraPhotoMapping) UnmarshalJSON(data []byte) error {
	m.initData()

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *CameraPhotoMapping
	if err := json.Unmarshal(data, tempType(m)); err != nil {
		return err
	}

	// Update parent references and keys.

	return nil
}
