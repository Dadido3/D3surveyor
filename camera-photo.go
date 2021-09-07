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
	"fmt"
	"image"
	"log"
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
	worldCoordinates := make([]Coordinate, 0, len(cp.Mappings))    // World coordinates of every point.
	imgCoordinates := make([]PixelCoordinate, 0, len(cp.Mappings)) // Image coordinates that each point is mapped to.
	for _, mapping := range cp.Mappings {
		if p, ok := site.Points[mapping.PointKey]; ok && !mapping.Suggested {
			mappings = append(mappings, mapping)
			worldCoordinates = append(worldCoordinates, p.Position.Coordinate)
			imgCoordinates = append(imgCoordinates, mapping.Position)
		}
	}

	// Unproject the mappings and project the points.
	//projectedCoordinates := cp.Project(worldCoordinates)      // The world coordinates of every point projected into image coordinates.
	unprojectedCoordinates, err := cp.Unproject(imgCoordinates) // The image coordinates of every mapped point projected into object/world coordinates.
	if err != nil {
		log.Printf("cp.UnProject() failed: %v", err)
		return 1000000
	}

	// Calculate the angle difference between rays going out the camera and points.
	// Sum up the squared angle residues.
	ssr := 0.0
	for i, mapping := range mappings {
		worldCoordinate, unprojectedCoordinate := worldCoordinates[i], unprojectedCoordinates[i]

		v1 := unprojectedCoordinate.Sub(cp.Position.Coordinate).Vec3()
		v2 := worldCoordinate.Sub(cp.Position.Coordinate).Vec3()

		angle := math.Acos(v1.Dot(v2) / v1.Len() / v2.Len())

		if math.IsNaN(angle) {
			ssr += 1000000
			continue
		}

		//pixelResidue := mgl64.Vec2{projectedCoordinate.X(), projectedCoordinate.Y()}.Sub(mgl64.Vec2{pointImg.X(), pointImg.Y()}).Len()

		// Square the weighted angular residue.
		r := (angle / float64(camera.AngAccuracy))
		sr := r * r
		sr = math.Min(sr, 1000000)
		mapping.sr = sr
		ssr += sr
	}

	return ssr
}

// Project transforms a list of object/world coordinates into a list of image coordinates.
func (cp *CameraPhoto) Project(worldCoordinates []Coordinate) []PixelCoordinate {
	projMatrix := cp.camera.GetProjectionMatrix(cp.ImageSize)
	viewMatrix := cp.GetViewMatrix()

	mvpMatrix := projMatrix.Mul4(viewMatrix)

	imgCoordinates := make([]PixelCoordinate, len(worldCoordinates))
	for i, worldCoordinate := range worldCoordinates {
		imgCoordinate := &imgCoordinates[i]

		obj4 := worldCoordinate.Vec4(1)

		vpp := mvpMatrix.Mul4x1(obj4)
		vpp = vpp.Mul(1 / vpp.W())

		imgCoordinate[0] = 0 + cp.ImageSize.X()*PixelDistance(vpp[0]+1)/2
		imgCoordinate[1] = cp.ImageSize.Y() - cp.ImageSize.Y()*PixelDistance(vpp[1]+1)/2
		imgCoordinate[2] = PixelDistance(vpp[2]+1) / 2
	}

	return imgCoordinates
}

// Unproject transforms a list of image coordinates into world coordinates.
func (cp *CameraPhoto) Unproject(imgCoordinates []PixelCoordinate) (worldCoordinates []Coordinate, err error) {
	projMatrix := cp.camera.GetProjectionMatrix(cp.ImageSize)
	viewMatrix := cp.GetViewMatrix()

	mvpMatrixInv := projMatrix.Mul4(viewMatrix).Inv()
	var blank mgl64.Mat4
	if mvpMatrixInv == blank {
		return nil, fmt.Errorf("could not find matrix inverse (projection times modelview is probably non-singular)")
	}

	worldCoordinates = make([]Coordinate, len(imgCoordinates))
	for i, imgCoordinate := range imgCoordinates {
		worldCoordinate := &worldCoordinates[i]

		obj4 := mvpMatrixInv.Mul4x1(mgl64.Vec4{
			float64((2*(imgCoordinate.X()-0))/cp.ImageSize.X() - 1),
			float64((2*-(imgCoordinate.Y()-cp.ImageSize.Y()))/cp.ImageSize.Y() - 1),
			float64(2*imgCoordinate.Z() - 1),
			1.0,
		})

		obj4[0] /= obj4[3]
		obj4[1] /= obj4[3]
		obj4[2] /= obj4[3]

		worldCoordinate[0] = Distance(obj4.X())
		worldCoordinate[1] = Distance(obj4.Y())
		worldCoordinate[2] = Distance(obj4.Z())
	}

	return worldCoordinates, nil
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
		if projectedCoordinate.Z() >= 0 && projectedCoordinate.Z() <= 1 && projectedCoordinate.X() >= 0 && projectedCoordinate.X() <= cp.ImageSize.X() && projectedCoordinate.Y() >= 0 && projectedCoordinate.Y() <= cp.ImageSize.Y() {
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

func (cp *CameraPhoto) GetViewMatrix() mgl64.Mat4 {
	quat := mgl64.AnglesToQuat(float64(-cp.Orientation.X()), float64(-cp.Orientation.Y()), float64(-cp.Orientation.Z()), mgl64.XYZ)
	rotationMatrix := quat.Mat4()
	translationMatrix := mgl64.Translate3D(float64(-cp.Position.X()), float64(-cp.Position.Y()), float64(-cp.Position.Z()))
	return rotationMatrix.Mul4(translationMatrix)
}
