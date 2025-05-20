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

	PointKey      string
	MappingKey    string // If non empty, we will only consider photos with this mapping.
	NavigationURL string // If non empty, the link this will navigate to upon clicking.

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
			var mapping *CameraPhotoMapping
			if c.MappingKey != "" {
				mapping = photo.Mappings[c.MappingKey]
			}

			if mapping == nil {
				// Fallback: Randomly pick some mapping for the given PointKey.
				for _, m := range photo.Mappings {
					if !m.Suggested && m.PointKey == c.PointKey {
						mapping = m
					}
				}
			}

			if mapping != nil {
				// Found mapping. Prepare all values for the UI.

				c.imgWidth, c.imgHeight = float64(photo.imageSize.X())*scaling, float64(photo.imageSize.Y())*scaling
				c.left, c.top = c.Width/2-float64(mapping.Position.X())*scaling, c.Height/2-float64(mapping.Position.Y())*scaling
				c.imageURL = photo.jsImageURL.String()

				return
			}
		}
	}

	c.imageURL = ""
}

func (c *PointViewComponent) handleClick(event vugu.DOMEvent) {
	switch {
	case c.NavigationURL != "":
		c.Navigate(c.NavigationURL, nil)
	case c.PointKey != "":
		c.Navigate("/point/"+c.PointKey, nil)
	case c.MappingKey != "":
		for _, camera := range c.Site.CamerasSorted() {
			for _, photo := range camera.Photos {
				if _, ok := photo.Mappings[c.MappingKey]; ok {
					c.Navigate("/camera/"+camera.Key()+"/photo/"+photo.Key(), nil)
					return
				}
			}
		}
	}
}
