package main

import (
	"github.com/vugu/vugu"
)

type PointViewComponent struct {
	Site *Site

	PointKey string
	imageURL string // The URL of the image as string.

	AttrMap vugu.AttrMap

	Width, Height       float64 // Width and height in DOM pixels.
	top, left           float64 // Image offset in DOM pixels.
	imgWidth, imgHeight float64 // Image width and height in DOM pixels.
}

func (pv *PointViewComponent) Compute(ctx vugu.ComputeCtx) {

	// Find camera that contains photo that contains the point we are looking for. Don't use suggested point mappings.
	for _, camera := range pv.Site.CamerasSorted() {
		for _, photo := range camera.Photos {
			for _, point := range photo.Points {
				if !point.Suggested && point.Point == pv.PointKey {
					// Found point. Set everything up.

					pv.imgWidth, pv.imgHeight = float64(photo.ImageWidth)/2, float64(photo.ImageHeight)/2
					pv.left, pv.top = pv.Width/2-point.X*pv.imgWidth, pv.Height/2-point.Y*pv.imgHeight
					pv.imageURL = photo.jsImageURL.String()

					return
				}
			}
		}
	}

	pv.imageURL = ""
}
