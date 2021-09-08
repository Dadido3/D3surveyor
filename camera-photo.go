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
	"bytes"
	"encoding/json"
	"image"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu/js"
)

type CameraPhoto struct {
	vgrouter.NavigatorRef `json:"-"`

	camera *Camera
	key    string

	CreatedAt time.Time

	ImageData          []byte // TODO: Don't store the image as byte slice. Only store it as a js blob
	ImageWidthMigrate  int    `json:"ImageWidth"`  // Value to be migrated. // TODO: Remove value migration in the future
	ImageHeightMigrate int    `json:"ImageHeight"` // Value to be migrated. // TODO: Remove value migration in the future
	ImageSize          PixelCoordinate

	jsImageBlob js.Value // Blob representing the image on the js side.
	jsImageURL  js.Value // URL referencing the blob.

	Position    CoordinateOptimizable
	Orientation RotationOptimizable

	Mappings map[string]*CameraPhotoMapping // List of mapped points.
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
		ImageSize: PixelCoordinate{PixelDistance(imageConf.Width), PixelDistance(imageConf.Height)},
		Mappings:  map[string]*CameraPhotoMapping{},
	}
	cp.createPhotoBlob(imageData)

	c.Photos[key] = cp

	return cp, nil
}

func (cp *CameraPhoto) Key() string {
	return cp.key
}

func (cp *CameraPhoto) Delete() {
	delete(cp.camera.Photos, cp.Key())

	if cp.jsImageURL.Truthy() {
		js.Global().Get("URL").Call("revokeObjectURL", cp.jsImageURL)
	}
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (cp *CameraPhoto) Copy() *CameraPhoto {
	copy := &CameraPhoto{
		CreatedAt:   cp.CreatedAt,
		ImageData:   cp.ImageData,
		ImageSize:   cp.ImageSize,
		jsImageBlob: cp.jsImageBlob,
		jsImageURL:  cp.jsImageURL,
		Position:    cp.Position,
		Orientation: cp.Orientation,
		Mappings:    map[string]*CameraPhotoMapping{},
	}

	// Generate copies of all children.
	for k, v := range cp.Mappings {
		copy.Mappings[k] = v.Copy()
	}

	// Restore keys and references.
	copy.RestoreChildrenRefs()

	return copy
}

// RestoreChildrenRefs updates the key of the children and any reference to this object.
func (cp *CameraPhoto) RestoreChildrenRefs() {
	for k, v := range cp.Mappings {
		v.key, v.photo = k, cp
	}
}

func (cp *CameraPhoto) UnmarshalJSON(data []byte) error {
	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *CameraPhoto
	if err := json.Unmarshal(data, tempType(cp)); err != nil {
		return err
	}

	// Load image on the js side.
	cp.createPhotoBlob(cp.ImageData)

	// Restore keys and references.
	cp.RestoreChildrenRefs()

	// Migrate some values.
	cp.ImageSize = PixelCoordinate{PixelDistance(cp.ImageWidthMigrate), PixelDistance(cp.ImageHeightMigrate)}

	for _, mapping := range cp.Mappings {
		mapping.Position = PixelCoordinate{
			PixelDistance(mapping.XMigrate) * cp.ImageSize.X(),
			PixelDistance(mapping.YMigrate) * cp.ImageSize.Y(),
		}
	}

	return nil
}

func (cp *CameraPhoto) createPhotoBlob(imageData []byte) {
	if cp.jsImageURL.Truthy() {
		js.Global().Get("URL").Call("revokeObjectURL", cp.jsImageURL)
	}

	dst := js.Global().Get("Uint8Array").New(len(imageData))
	js.CopyBytesToJS(dst, imageData)
	dstArray := js.Global().Get("Array").New(dst)

	cp.jsImageBlob = js.Global().Get("Blob").New(dstArray, js.ValueOf(map[string]interface{}{"type": "image/*"}))
	cp.jsImageURL = js.Global().Get("URL").Call("createObjectURL", cp.jsImageBlob) // This has to be freed when the photo is deleted.
}

func (cp *CameraPhoto) JsImageURL() string {
	return cp.jsImageURL.String()
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (cp *CameraPhoto) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables1, _ := cp.Position.GetTweakablesAndResiduals()
	tweakables2, _ := cp.Orientation.GetTweakablesAndResiduals()

	return append(append([]Tweakable{}, tweakables1...), tweakables2...), []Residualer{cp}
}

// ResidualSqr returns the sum of squared residuals. (Each residual is divided by the accuracy of the measurement device).
func (cp *CameraPhoto) ResidualSqr() float64 {
	camera, site := cp.camera, cp.camera.site

	// Generate list of mappings and their coordinates. Ignore suggested mappings.
	mappings := make([]*CameraPhotoMapping, 0, len(cp.Mappings))
	worldCoordinates := make([]Coordinate, 0, len(cp.Mappings)) // World coordinates of every point.
	for _, mapping := range cp.Mappings {
		if p, ok := site.Points[mapping.PointKey]; ok && !mapping.Suggested {
			mappings = append(mappings, mapping)
			worldCoordinates = append(worldCoordinates, p.Position.Coordinate)
		}
	}

	// Project the points.
	projectedCoordinates := cp.Project(worldCoordinates) // The world coordinates of every point projected into image coordinates.

	// Calculate the distance in pixels for every point.
	// Sum up the squared distance residues.
	ssr := 0.0
	for i, mapping := range mappings {
		projectedCoordinate := projectedCoordinates[i]

		// Ignore points behind the photo.
		if projectedCoordinate.Z() <= 0 {
			ssr += 1000000
			continue
		}

		sr := projectedCoordinate.DistanceSqr(mapping.Position) / camera.PixelAccuracy.Sqr()
		sr = math.Min(sr, 1000000)
		mapping.sr = sr
		ssr += sr
	}

	return ssr
}

// Project transforms a list of object/world coordinates into a list of (distorted) image coordinates.
func (cp *CameraPhoto) Project(worldCoordinates []Coordinate) []PixelCoordinate {
	camera := cp.camera

	imageCenter := cp.ImageSize.Scaled(0.5)

	focalLength := imageCenter.X().Pixels() / math.Tan(camera.HorizontalAOV.Radian()/2)

	cameraMatrix := cp.GetCameraViewMatrix()

	projectedCoordinates := make([]PixelCoordinate, len(worldCoordinates))
	for i, worldCoordinate := range worldCoordinates {
		obj4 := worldCoordinate.Vec4(1)

		// Rotate and translate the world coordinate into the camera coordinate system.
		loc4 := cameraMatrix.Mul4x1(obj4)

		// Scale X and Y camera coordinates on Z distance. (Perspective projection)
		localCoordinate := PixelCoordinate{
			PixelDistance(loc4[0] / loc4[2]),
			PixelDistance(loc4[1] / loc4[2]),
			PixelDistance(loc4[2]),
		}

		// Radially distort the coordinates.
		radiusSqr := localCoordinate.LengthSqr()
		distortedCoordinate := localCoordinate.Scaled(camera.radialDistortFactor(radiusSqr))

		projectedCoordinates[i] = imageCenter.Add(camera.DistortionCenterOffset).Add(distortedCoordinate.Scaled(focalLength))
	}

	return projectedCoordinates
}

// UpdateSuggestions recreates/updates all "suggested" point mappings.
func (cp *CameraPhoto) UpdateSuggestions() {
	site := cp.camera.site

	// Get a list of all points and their world coordinates.
	points := make([]*Point, 0, len(site.Points))
	worldCoordinates := make([]Coordinate, 0, len(site.Points)) // World coordinates of every point.
	for _, point := range site.Points {
		points = append(points, point)
		worldCoordinates = append(worldCoordinates, point.Position.Coordinate)
	}

	// Project the points onto the photo.
	projectedCoordinates := cp.Project(worldCoordinates) // The object/world coordinates of every point projected into image coordinates.

	// Update suggested mapped points.
	for i, projectedCoordinate := range projectedCoordinates {
		point := points[i]
		if projectedCoordinate.Z() > 0 && projectedCoordinate.X() >= 0 && projectedCoordinate.X() <= cp.ImageSize.X() && projectedCoordinate.Y() >= 0 && projectedCoordinate.Y() <= cp.ImageSize.Y() {
			// The projection is valid, create or update point mapping.
			var foundMapping *CameraPhotoMapping
			for _, mapping := range cp.Mappings { // TODO: Remove stupid linear search
				if mapping.PointKey == point.Key() {
					foundMapping = mapping
					break
				}
			}
			if foundMapping == nil {
				foundMapping = cp.NewMapping()
				foundMapping.Suggested = true
			}
			// Update the projected coordinate for every found point.
			foundMapping.projectedPos = projectedCoordinate
			// Only update suggested point mappings, not user placed ones.
			if foundMapping.Suggested {
				foundMapping.Position, foundMapping.PointKey = projectedCoordinate, point.Key()
			}
		} else {
			// The projection is outside of the image, remove the suggested point mapping if there is any.
			for _, mapping := range cp.Mappings { // TODO: Remove stupid linear search
				if mapping.PointKey == point.Key() {
					if mapping.Suggested {
						// Delete if it's a suggested mapping.
						mapping.Delete()
					} else {
						// Otherwise just set the projected coordinate.
						mapping.projectedPos = projectedCoordinate
					}
					break
				}
			}
		}
	}

	// Remove any suggested mappings to non existing points.
	for _, mapping := range cp.Mappings {
		if mapping.Suggested {
			if _, found := site.Points[mapping.PointKey]; !found {
				mapping.Delete()
			}
		}
	}
}

// GetCameraViewMatrix returns a matrix that transforms world coordinates into local camera coordinates.
func (cp *CameraPhoto) GetCameraViewMatrix() mgl64.Mat4 {
	quat := mgl64.AnglesToQuat(float64(-cp.Orientation.X()), float64(-cp.Orientation.Y()), float64(-cp.Orientation.Z()), mgl64.XYZ)
	rotationMatrix := quat.Mat4()
	translationMatrix := mgl64.Translate3D(float64(-cp.Position.X()), float64(-cp.Position.Y()), float64(-cp.Position.Z()))
	return rotationMatrix.Mul4(translationMatrix)
}
