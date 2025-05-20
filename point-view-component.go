// Copyright (C) 2021-2025 David Vogel
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
	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
)

type PointViewComponent struct {
	vgrouter.NavigatorRef `json:"-"`

	Site *Site

	PointKey string
	imageURL string // The URL of the image as string.

	AttrMap vugu.AttrMap

	Width, Height       float64 // Width and height in DOM pixels.
	Scale               float64 // The scale of the image. Larger values scale the image up. Defaults to 1 when not set.
	top, left           float64 // Image offset in DOM pixels.
	imgWidth, imgHeight float64 // Image width and height in DOM pixels.
}

func (c *PointViewComponent) Compute(ctx vugu.ComputeCtx) {

	scaling := 1.0
	if c.Scale >= 0 {
		scaling = c.Scale
	}

	// Find camera that contains photo that contains the point we are looking for. Don't use suggested point mappings.
	for _, camera := range c.Site.CamerasSorted() {
		for _, photo := range camera.Photos {
			for _, mapping := range photo.Mappings {
				if !mapping.Suggested && mapping.PointKey == c.PointKey {
					// Found point. Set everything up.

					c.imgWidth, c.imgHeight = float64(photo.imageSize.X())*scaling, float64(photo.imageSize.Y())*scaling
					c.left, c.top = c.Width/2-float64(mapping.Position.X())*scaling, c.Height/2-float64(mapping.Position.Y())*scaling
					c.imageURL = photo.jsImageURL.String()

					return
				}
			}
		}
	}

	c.imageURL = ""
}
