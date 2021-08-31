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
				if !point.Suggested && point.PointKey == pv.PointKey {
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
