package main

import (
	"bytes"
	"encoding/json"
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

	ImageData               []byte // TODO: Don't store the image as byte slice. Only store it as a js blob.
	ImageWidth, ImageHeight int

	jsImageBlob js.Value // Blob representing the image on the js side.
	jsImageURL  js.Value // URL referencing the blob.

	Position    Coordinate
	Orientation Rotation

	// Projection matrix computed from the position and orientation.
	projMatrix mgl64.Mat4

	Points map[string]*CameraPhotoPoint // List of points mapped to this photo.
}

func (c *Camera) NewPhoto(imageData []byte) (*CameraPhoto, error) {
	imageConf, _, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	key := c.site.shortIDGen.MustGenerate()

	cp := &CameraPhoto{
		camera:      c,
		key:         key,
		CreatedAt:   time.Now(),
		ImageData:   imageData,
		ImageWidth:  imageConf.Width,
		ImageHeight: imageConf.Height,
		Points:      map[string]*CameraPhotoPoint{},
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

func (cp *CameraPhoto) UnmarshalJSON(data []byte) error {
	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *CameraPhoto
	if err := json.Unmarshal(data, tempType(cp)); err != nil {
		return err
	}

	// Load image on the js side.
	cp.createPhotoBlob(cp.ImageData)

	// Restore keys and references.
	for k, v := range cp.Points {
		v.key, v.photo = k, cp
	}

	return nil
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

	width, height := cp.ImageWidth, cp.ImageHeight
	aspect := float64(width) / float64(height)

	var fovy float64
	if width > height {
		fovy = float64(camera.LongSideFOV) / aspect
	} else {
		fovy = float64(camera.LongSideFOV)
	}

	cp.projMatrix = mgl64.Perspective(fovy, aspect, 0.001, 1)

	quat := mgl64.AnglesToQuat(float64(-cp.Orientation.X), float64(-cp.Orientation.Y), float64(-cp.Orientation.Z), mgl64.XYZ)
	rotationMatrix := quat.Mat4()
	translationMatrix := mgl64.Translate3D(float64(-cp.Position.X), float64(-cp.Position.Y), float64(-cp.Position.Z))
	viewMatrix := rotationMatrix.Mul4(translationMatrix)

	// Calculate the angle difference between rays going out the camera and points.
	// Sum up the squared angle residues.
	ssr := 0.0
	for _, point := range cp.Points {
		if p, ok := site.Points[point.Point]; ok {
			v, err := mgl64.UnProject(mgl64.Vec3{point.X, point.Y, 1}, viewMatrix, cp.projMatrix, 0, 1, 1, -1)
			if err != nil {
				log.Printf("mgl64.UnProject() failed: %v", err)
				continue
			}

			// TODO: Put calculation of the real coordinate somewhere else
			realV := mgl64.Project(mgl64.Vec3{float64(p.Position.X), float64(p.Position.Y), float64(p.Position.Z)}, viewMatrix, cp.projMatrix, 0, 1, 1, -1)
			point.realX, point.realY = realV.X(), realV.Y()

			v1 := mgl64.Vec3{float64(p.Position.X - cp.Position.X), float64(p.Position.Y - cp.Position.Y), float64(p.Position.Z - cp.Position.Z)}
			v2 := mgl64.Vec3{v.X() - float64(cp.Position.X), v.Y() - float64(cp.Position.Y), v.Z() - float64(cp.Position.Z)}

			angle := math.Acos(v1.Dot(v2) / v1.Len() / v2.Len())

			//fmt.Println(cp.Key(), v, angle)

			if math.IsNaN(angle) {
				continue
			}

			pixelResidue := mgl64.Vec2{point.realX * float64(width), point.realY * float64(height)}.Sub(mgl64.Vec2{point.X * float64(width), point.Y * float64(height)}).Len()

			// The residue that gets squared is a mix of the angluar residue divided by the angular accuracy, and the pixel residue divided by a hard coded accuracy of 5 pixels.
			sr := math.Pow(angle/float64(camera.AngAccuracy), 2) + math.Pow(pixelResidue/5, 2)
			point.sr = sr
			ssr += sr
		}
	}

	return ssr
}
