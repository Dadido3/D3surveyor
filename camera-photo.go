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

	ImageData               []byte // TODO: Don't store the image as byte slice. Only store it as a js blob.
	ImageWidth, ImageHeight int

	jsImageBlob js.Value // Blob representing the image on the js side.
	jsImageURL  js.Value // URL referencing the blob.

	Position    Coordinate
	Orientation Rotation

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

	// Generate list of mapped points and their coordinates.
	points := make([]*CameraPhotoPoint, 0, len(cp.Points))
	pointsObj := make([]mgl64.Vec3, 0, len(cp.Points)) // Object/World coordinates of every point.
	pointsImg := make([]mgl64.Vec3, 0, len(cp.Points)) // Image coordinates where every point is mapped to.
	for _, point := range cp.Points {
		if p, ok := site.Points[point.Point]; ok {
			points = append(points, point)
			pointsObj = append(pointsObj, mgl64.Vec3{float64(p.Position.X), float64(p.Position.Y), float64(p.Position.Z)})
			pointsImg = append(pointsImg, mgl64.Vec3{point.X, point.Y, 1})
		}
	}

	// Project and unproject the points.
	pointsProjected := cp.Project(pointsObj)          // The object/world coordinates of every point projected into image coordinates.
	pointsUnprojected, err := cp.UnProject(pointsImg) // The mapped image coordinates of every point projected into object/world coordinates.
	if err != nil {
		log.Printf("cp.UnProject() failed: %v", err)
		return 1000000
	}

	// Calculate the angle difference between rays going out the camera and points.
	// Sum up the squared angle residues.
	ssr := 0.0
	for i, point := range points {
		pointObj, pointImg, pointUnprojected, pointProjected := pointsObj[i], pointsImg[i], pointsUnprojected[i], pointsProjected[i]

		// TODO: Put calculation of the real coordinate somewhere else
		point.projectedX, point.projectedY = pointProjected.X(), pointProjected.Y()

		cpCoordinate := mgl64.Vec3{float64(cp.Position.X), float64(cp.Position.Y), float64(cp.Position.Z)}
		v1 := pointUnprojected.Sub(cpCoordinate)
		v2 := pointObj.Sub(cpCoordinate)

		angle := math.Acos(v1.Dot(v2) / v1.Len() / v2.Len())

		if math.IsNaN(angle) {
			ssr += 1000000
			continue
		}

		pixelResidue := mgl64.Vec2{pointProjected.X() * float64(width), pointProjected.Y() * float64(height)}.Sub(mgl64.Vec2{pointImg.X() * float64(width), pointImg.Y() * float64(height)}).Len()

		// The residue that gets squared is a mix of the angluar residue divided by the angular accuracy, and the pixel residue divided by a hard coded accuracy of 50 pixels.
		sr := math.Pow(angle/float64(camera.AngAccuracy), 2) + math.Pow(pixelResidue/50, 2)
		sr = math.Min(sr, 1000000)
		point.sr = sr
		ssr += sr
	}

	return ssr
}

// Project transforms a list of object/world coordinates into a list of image coordinates [0,1].
func (cp *CameraPhoto) Project(objs []mgl64.Vec3) (win []mgl64.Vec3) {
	projMatrix := cp.camera.GetProjectionMatrix(float64(cp.ImageWidth), float64(cp.ImageHeight))
	viewMatrix := cp.GetViewMatrix()

	mvpMatrix := projMatrix.Mul4(viewMatrix)

	wins := make([]mgl64.Vec3, len(objs))
	for i, obj := range objs {
		win := &wins[i]

		obj4 := obj.Vec4(1)

		vpp := mvpMatrix.Mul4x1(obj4)
		vpp = vpp.Mul(1 / vpp.W())
		win[0] = 0 + (vpp[0]+1)/2
		win[1] = 1 - (vpp[1]+1)/2
		win[2] = (vpp[2] + 1) / 2
	}

	return wins
}

// UnProject transforms a list of image coordinates into object/world coordinates.
func (cp *CameraPhoto) UnProject(wins []mgl64.Vec3) (obj []mgl64.Vec3, err error) {
	projMatrix := cp.camera.GetProjectionMatrix(float64(cp.ImageWidth), float64(cp.ImageHeight))
	viewMatrix := cp.GetViewMatrix()

	mvpMatrixInv := projMatrix.Mul4(viewMatrix).Inv()
	var blank mgl64.Mat4
	if mvpMatrixInv == blank {
		return nil, fmt.Errorf("could not find matrix inverse (projection times modelview is probably non-singular)")
	}

	objs := make([]mgl64.Vec3, len(wins))
	for i, win := range wins {
		obj := &objs[i]

		obj4 := mvpMatrixInv.Mul4x1(mgl64.Vec4{
			(2 * (win[0] - 0)) - 1,
			(2 * -(win[1] - 1)) - 1,
			2*win[2] - 1,
			1.0,
		})
		*obj = obj4.Vec3()

		obj[0] /= obj4[3]
		obj[1] /= obj4[3]
		obj[2] /= obj4[3]
	}

	return objs, nil
}

func (cp *CameraPhoto) GetViewMatrix() mgl64.Mat4 {
	quat := mgl64.AnglesToQuat(float64(-cp.Orientation.X), float64(-cp.Orientation.Y), float64(-cp.Orientation.Z), mgl64.XYZ)
	rotationMatrix := quat.Mat4()
	translationMatrix := mgl64.Translate3D(float64(-cp.Position.X), float64(-cp.Position.Y), float64(-cp.Position.Z))
	return rotationMatrix.Mul4(translationMatrix)
}
