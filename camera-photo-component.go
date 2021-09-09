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
	"fmt"
	"math"

	"github.com/vugu/vugu"
	js "github.com/vugu/vugu/js"
)

type CameraPhotoComponentEventCoordinate struct {
	xCan, yCan PixelDistance // Position in canvas coordinates.
}

type CameraPhotoComponent struct {
	Photo *CameraPhoto

	// Canvas state variables. // TODO: Put most of the "scrollable/zoomable canvas" logic in its own module for reusability
	scale                     float64       // Ratio between Canvas and virtual coordinates.
	originX, originY          PixelDistance // Origin in canvas coordinates. // TODO: Use specific type for DOM, Canvas and virtual coordinates, so that you can't mix them up
	canWidth, canHeight       PixelDistance // Width and height in canvas pixels.
	canWidthDOM, canHeightDOM PixelDistance // Width and height of the canvas in dom pixels.

	cachedImg js.Value // Cached js image object.

	showLines, showRangefinders, showTripods bool

	ongoingMouseDrags map[int]CameraPhotoComponentEventCoordinate
	ongoingTouches    map[int]CameraPhotoComponentEventCoordinate

	selectedMapping    *CameraPhotoMapping
	highlightedMapping *CameraPhotoMapping
}

func (c *CameraPhotoComponent) canvasCreated(canvas js.Value) {
	// TODO: Put this into a resize event or something similar
	//c.canWidth, c.canHeight = canvas.Get("width").Float(), canvas.Get("height").Float()

	rect := canvas.Call("getBoundingClientRect")
	c.canWidthDOM, c.canHeightDOM = PixelDistance(rect.Get("width").Float()), PixelDistance(rect.Get("height").Float())

	c.canWidth, c.canHeight = c.canWidthDOM, c.canHeightDOM
	canvas.Set("width", c.canWidth.Pixels())
	canvas.Set("height", c.canHeight.Pixels())

	c.canvasRedraw(canvas)
}

func (c *CameraPhotoComponent) Init(ctx vugu.InitCtx) {
	if c.ongoingMouseDrags == nil {
		c.ongoingMouseDrags = make(map[int]CameraPhotoComponentEventCoordinate)
	}
	if c.ongoingTouches == nil {
		c.ongoingTouches = make(map[int]CameraPhotoComponentEventCoordinate)
	}

	c.showLines, c.showRangefinders = true, false

	if c.scale == 0 {
		c.setScale(1, 0, 0)
	}
}

func (c *CameraPhotoComponent) handleUnmap(event vugu.DOMEvent) {
	for _, mapping := range c.Photo.Mappings {
		if mapping == c.selectedMapping {
			c.selectedMapping = nil
			mapping.Delete()
			return
		}
	}
}

func (c *CameraPhotoComponent) handleContextMenu(event vugu.DOMEvent) {
	//jsEvent, jsCanvas := event.JSEvent(), event.JSEventTarget()

	//jsEvent.Call("preventDefault")
	//jsEvent.Call("stopPropagation")
}

func (c *CameraPhotoComponent) handlePointerDown(event vugu.DOMEvent) {
	jsEvent, jsCanvas := event.JSEvent(), event.JSEventTarget()

	//jsEvent.Call("preventDefault")
	//jsEvent.Call("stopPropagation")

	pointerID := jsEvent.Get("pointerId").Int()
	inputType := jsEvent.Get("pointerType").String()
	jsCanvas.Call("setPointerCapture", pointerID)
	xCan, yCan := c.transformDOMToCanvas(PixelDistance(jsEvent.Get("offsetX").Float()), PixelDistance(jsEvent.Get("offsetY").Float()))
	switch inputType {
	case "mouse":
		// Store current pointer.
		c.ongoingMouseDrags[pointerID] = CameraPhotoComponentEventCoordinate{
			xCan: xCan,
			yCan: yCan,
		}
		// Check if current selection is still the one under the pointer.
		mapping, _, _ := c.getClosestMapping(xCan, yCan, 20*20)
		if c.selectedMapping != mapping {
			c.selectedMapping = nil
		}

	case "touch":
		// Store current pointer.
		c.ongoingTouches[pointerID] = CameraPhotoComponentEventCoordinate{
			xCan: xCan,
			yCan: yCan,
		}
		// Check if current selection is still the one under the pointer.
		mapping, _, _ := c.getClosestMapping(xCan, yCan, 20*20)
		if c.selectedMapping != mapping {
			c.selectedMapping = nil
		}

	default:
		fmt.Printf("Input type %q not supported\n", inputType)
	}
}

func (c *CameraPhotoComponent) handlePointerMove(event vugu.DOMEvent) {
	jsEvent, _ := event.JSEvent(), event.JSEventTarget()

	//jsEvent.Call("preventDefault")
	//jsEvent.Call("stopPropagation")

	pointerID := jsEvent.Get("pointerId").Int()
	inputType := jsEvent.Get("pointerType").String()
	xCan, yCan := c.transformDOMToCanvas(PixelDistance(jsEvent.Get("offsetX").Float()), PixelDistance(jsEvent.Get("offsetY").Float()))
	switch inputType {
	case "mouse":
		if ongoingMouseDrag, ok := c.ongoingMouseDrags[pointerID]; ok {
			if c.selectedMapping != nil {
				// Drag selected point.
				c.selectedMapping.Position = c.selectedMapping.Position.Add(PixelCoordinate{
					(xCan - ongoingMouseDrag.xCan) / PixelDistance(c.scale),
					(yCan - ongoingMouseDrag.yCan) / PixelDistance(c.scale),
				})
			} else {
				// Drag viewport.
				c.originX += xCan - ongoingMouseDrag.xCan
				c.originY += yCan - ongoingMouseDrag.yCan
			}

			c.ongoingMouseDrags[pointerID] = CameraPhotoComponentEventCoordinate{
				xCan: xCan,
				yCan: yCan,
			}
		}

		// Get highlighted element.
		mapping, _, _ := c.getClosestMapping(xCan, yCan, 20*20)
		if c.highlightedMapping != mapping {
			c.highlightedMapping = mapping
		}

	case "touch":
		switch len(c.ongoingTouches) {
		case 1:
			if ongoingTouch, ok := c.ongoingTouches[pointerID]; ok {
				if c.selectedMapping != nil {
					// Drag selected point.
					c.selectedMapping.Position = c.selectedMapping.Position.Add(PixelCoordinate{
						(xCan - ongoingTouch.xCan) / PixelDistance(c.scale),
						(yCan - ongoingTouch.yCan) / PixelDistance(c.scale),
					})
				} else {
					// Drag viewport.
					c.originX += xCan - ongoingTouch.xCan
					c.originY += yCan - ongoingTouch.yCan
				}

				c.ongoingTouches[pointerID] = CameraPhotoComponentEventCoordinate{
					xCan: xCan,
					yCan: yCan,
				}
			}
		case 2:
			if ongoingTouch, ok := c.ongoingTouches[pointerID]; ok {

				// Get the other touch event, scale the viewport accordingly to the change in finger distance.
				for otherPointerID, ongoingOtherTouch := range c.ongoingTouches {
					if pointerID != otherPointerID {
						prevDistance := math.Sqrt((ongoingTouch.xCan - ongoingOtherTouch.xCan).Sqr() + (ongoingTouch.yCan - ongoingOtherTouch.yCan).Sqr())
						newDistance := math.Sqrt((xCan - ongoingOtherTouch.xCan).Sqr() + (yCan - ongoingOtherTouch.yCan).Sqr())
						pivotX, pivotY := c.transformCanvasToVirtual((ongoingTouch.xCan+ongoingOtherTouch.xCan)/2, (ongoingTouch.yCan+ongoingOtherTouch.yCan)/2)
						c.setScale(c.scale*float64(newDistance/prevDistance), pivotX, pivotY)
						break
					}
				}

				// Translate view along the common middlepoint of the two touch events.
				c.originX += (xCan - ongoingTouch.xCan) / 2
				c.originY += (yCan - ongoingTouch.yCan) / 2

				c.ongoingTouches[pointerID] = CameraPhotoComponentEventCoordinate{
					xCan: xCan,
					yCan: yCan,
				}
			}
		}

	default:
		fmt.Printf("Input type %q not supported\n", inputType)
	}
}

func (c *CameraPhotoComponent) handlePointerUp(event vugu.DOMEvent) {
	jsEvent, jsCanvas := event.JSEvent(), event.JSEventTarget()

	//jsEvent.Call("preventDefault")
	//jsEvent.Call("stopPropagation")

	pointerID := jsEvent.Get("pointerId").Int()
	inputType := jsEvent.Get("pointerType").String()
	xCan, yCan := c.transformDOMToCanvas(PixelDistance(jsEvent.Get("offsetX").Float()), PixelDistance(jsEvent.Get("offsetY").Float()))
	switch inputType {
	case "mouse":
		jsCanvas.Call("releasePointerCapture", pointerID)
		// Reset highlighted point.
		c.highlightedMapping = nil
		// Get element selection.
		mapping, _, _ := c.getClosestMapping(xCan, yCan, 20*20)
		if c.selectedMapping != mapping {
			c.selectedMapping = mapping
		}
		delete(c.ongoingMouseDrags, pointerID)

	case "touch":
		jsCanvas.Call("releasePointerCapture", pointerID)

		// Get element selection.
		mapping, _, _ := c.getClosestMapping(xCan, yCan, 20*20)
		if c.selectedMapping != mapping {
			c.selectedMapping = mapping
		}

		delete(c.ongoingTouches, pointerID)

	default:
		fmt.Printf("Input type %q not supported\n", inputType)
	}
}

func (c *CameraPhotoComponent) handleDblClick(event vugu.DOMEvent) {
	jsEvent, _ := event.JSEvent(), event.JSEventTarget()

	xCan, yCan := c.transformDOMToCanvas(PixelDistance(jsEvent.Get("offsetX").Float()), PixelDistance(jsEvent.Get("offsetY").Float()))
	xVir, yVir := c.transformCanvasToVirtual(xCan, yCan)

	mapping, _, _ := c.getClosestMapping(xCan, yCan, 20*20)

	if mapping != nil {
		// Convert suggested point mapping in user created one.
		mapping.Suggested = false
	} else {
		// Create new mapping mapping at event position.
		mapping := c.Photo.NewMapping()
		mapping.Position = PixelCoordinate{xVir, yVir}
	}
}

func (c *CameraPhotoComponent) handleClick(event vugu.DOMEvent) {

}

func (c *CameraPhotoComponent) handleWheel(event vugu.DOMEvent) {
	jsEvent, _ := event.JSEvent(), event.JSEventTarget()

	jsEvent.Call("preventDefault")
	jsEvent.Call("stopPropagation")

	delta := jsEvent.Get("deltaY").Float()
	xPivotCan, yPivotCan := c.transformDOMToCanvas(PixelDistance(jsEvent.Get("offsetX").Float()), PixelDistance(jsEvent.Get("offsetY").Float()))
	xPivot, yPivot := c.transformCanvasToVirtual(xPivotCan, yPivotCan)

	c.setScale(math.Pow(1/1.001, delta)*c.scale, xPivot, yPivot)
}

// transformDOMToCanvas takes the coordinates relative to the top left of the element in DOM pixels and transforms them into the canvas coordinates.
func (c *CameraPhotoComponent) transformDOMToCanvas(xDOM, yDOM PixelDistance) (xCan, yCan PixelDistance) {
	return xDOM / c.canWidthDOM * c.canWidth, yDOM / c.canHeightDOM * c.canHeight
}

// transformCanvasToVirtual takes canvas coordinates and transforms them into virtual coordinates.
func (c *CameraPhotoComponent) transformCanvasToVirtual(xCan, yCan PixelDistance) (xVir, yVir PixelDistance) {
	return (xCan - c.originX) / PixelDistance(c.scale), (yCan - c.originY) / PixelDistance(c.scale)
}

// transformVirtualToCanvas takes virtual coordinates and transforms them into canvas coordinates.
func (c *CameraPhotoComponent) transformVirtualToCanvas(xVir, yVir PixelDistance) (xCan, yCan PixelDistance) {
	return xVir*PixelDistance(c.scale) + c.originX, yVir*PixelDistance(c.scale) + c.originY
}

// setScale clamps and overwrites the current scale.
// xPivot and yPivot are in virtual coordinates.
func (c *CameraPhotoComponent) setScale(newScale float64, xPivot, yPivot PixelDistance) {
	// Clamp new scale.
	if newScale > 10 {
		newScale = 10
	} else if newScale < 0.01 {
		newScale = 0.01
	}

	// Calculate new origin in canvas coordinates.
	c.originX, c.originY = c.originX+xPivot*PixelDistance(c.scale-newScale), c.originY+yPivot*PixelDistance(c.scale-newScale)

	c.scale = newScale
}

// transformUnscaled sets the transformation matrix to have its origin at the given virtual coordinate, but the scale is set to 1.
func (c *CameraPhotoComponent) transformUnscaled(drawCtx js.Value, xVir, yVir PixelDistance) {
	xCan, yCan := c.transformVirtualToCanvas(xVir, yVir)
	drawCtx.Call("setTransform", 1, 0, 0, 1, xCan.Pixels(), yCan.Pixels())
}

// transformScaled sets the transformation matrix to represent the origin and scaling values.
func (c *CameraPhotoComponent) transformScaled(drawCtx js.Value) {
	drawCtx.Call("setTransform", c.scale, 0, 0, c.scale, c.originX.Pixels(), c.originY.Pixels())
}

// getClosestMapping returns the closest mapped point to the given canvas coordinates.
func (c *CameraPhotoComponent) getClosestMapping(xCan, yCan PixelDistance, maxDistSqr float64) (minMapping *CameraPhotoMapping, minKey string, minDistSqr float64) {
	minDistSqr = maxDistSqr

	for key, mapping := range c.Photo.Mappings {
		pXCan, pYCan := c.transformVirtualToCanvas(mapping.Position.X(), mapping.Position.Y())
		distSqr := (pXCan - xCan).Sqr() + (pYCan - yCan).Sqr()
		if minDistSqr > distSqr {
			minDistSqr, minKey, minMapping = distSqr, key, mapping
		}
	}

	return
}

func (c *CameraPhotoComponent) canvasRedraw(canvas js.Value) {

	site := c.Photo.camera.site

	drawCtx := canvas.Call("getContext", "2d")

	if c.cachedImg.IsUndefined() {
		c.cachedImg = js.Global().Get("Image").New()
		c.cachedImg.Set("src", c.Photo.jsImageURL)
	}

	// Recalculate suggested mappings and projected coordinates.
	c.Photo.UpdateSuggestions() // TODO: Recalculate suggested point mappings more intelligent

	drawCtx.Set("shadowBlur", 0)

	drawCtx.Call("setTransform", 1, 0, 0, 1, 0, 0)
	drawCtx.Call("clearRect", 0, 0, c.canWidth.Pixels(), c.canHeight.Pixels())
	c.transformScaled(drawCtx)

	drawCtx.Call("drawImage", c.cachedImg, 0, 0)

	if c.showLines {
		drawCtx.Set("lineWidth", 1)
		drawCtx.Set("lineCap", "butt")
		drawCtx.Set("strokeStyle", "blue")
		drawCtx.Call("setLineDash", []interface{}{})
		drawCtx.Set("shadowBlur", 0)
		for _, line := range site.Lines {
			p1, p2 := line.P1, line.P2

			var foundM1, foundM2 *CameraPhotoMapping
			for _, mapping := range c.Photo.Mappings {
				if mapping.PointKey == "" {
					continue
				}
				if mapping.PointKey == p1 {
					foundM1 = mapping
				}
				if mapping.PointKey == p2 {
					foundM2 = mapping
				}

				if foundM1 != nil && foundM2 != nil {
					break
				}
			}

			if foundM1 != nil && foundM2 != nil {
				c.transformUnscaled(drawCtx, foundM1.Position.X(), foundM1.Position.Y())
				drawCtx.Call("beginPath")
				drawCtx.Call("moveTo", 0, 0)
				c.transformUnscaled(drawCtx, foundM2.Position.X(), foundM2.Position.Y())
				drawCtx.Call("lineTo", 0, 0)
				drawCtx.Call("stroke")
			}
		}
	}

	if c.showRangefinders {
		drawCtx.Set("lineWidth", 2)
		drawCtx.Set("lineCap", "butt")
		drawCtx.Set("strokeStyle", "yellow")
		drawCtx.Call("setLineDash", []interface{}{5, 10})
		drawCtx.Set("shadowBlur", 0)
		for _, rangefinder := range site.Rangefinders {
			for _, measurement := range rangefinder.Measurements {
				p1, p2 := measurement.P1, measurement.P2

				var foundM1, foundM2 *CameraPhotoMapping
				for _, mapping := range c.Photo.Mappings {
					if mapping.PointKey == "" {
						continue
					}
					if mapping.PointKey == p1 {
						foundM1 = mapping
					}
					if mapping.PointKey == p2 {
						foundM2 = mapping
					}

					if foundM1 != nil && foundM2 != nil {
						break
					}
				}

				if foundM1 != nil && foundM2 != nil {
					c.transformUnscaled(drawCtx, foundM1.Position.X(), foundM1.Position.Y())
					drawCtx.Call("beginPath")
					drawCtx.Call("moveTo", 0, 0)
					c.transformUnscaled(drawCtx, foundM2.Position.X(), foundM2.Position.Y())
					drawCtx.Call("lineTo", 0, 0)
					drawCtx.Call("stroke")
				}

			}
		}
	}

	if c.showTripods {
		drawCtx.Set("lineWidth", 1)
		drawCtx.Set("lineCap", "butt")
		drawCtx.Set("strokeStyle", "green")
		drawCtx.Call("setLineDash", []interface{}{5, 2})
		drawCtx.Set("shadowBlur", 0)
		for _, tripod := range site.Tripods {
			tripodProjectedTemp, _ := c.Photo.Project([]Coordinate{tripod.Position.Coordinate})
			tripodProjected := tripodProjectedTemp[0] // TODO: Filter out tripods that can't be really projected

			for _, measurement := range tripod.Measurements {
				pointKey := measurement.PointKey

				var foundMapping *CameraPhotoMapping
				for _, mapping := range c.Photo.Mappings {
					if mapping.PointKey == "" {
						continue
					}
					if mapping.PointKey == pointKey {
						foundMapping = mapping
						break
					}
				}

				if foundMapping != nil {
					c.transformUnscaled(drawCtx, foundMapping.Position.X(), foundMapping.Position.Y())
					drawCtx.Call("beginPath")
					drawCtx.Call("moveTo", 0, 0)
					c.transformUnscaled(drawCtx, tripodProjected.X(), tripodProjected.Y())
					drawCtx.Call("lineTo", 0, 0)
					drawCtx.Call("stroke")
				}

			}
		}
	}

	drawCtx.Set("lineWidth", 1)
	drawCtx.Set("lineCap", "butt")
	drawCtx.Call("setLineDash", []interface{}{})
	drawCtx.Set("shadowOffsetX", 0)
	drawCtx.Set("shadowOffsetY", 0)
	drawCtx.Set("shadowColor", "white")
	for _, mapping := range c.Photo.Mappings {
		point, pointOk := site.Points[mapping.PointKey]

		if mapping == c.selectedMapping {
			drawCtx.Set("strokeStyle", "white")
		} else if mapping == c.highlightedMapping {
			drawCtx.Set("strokeStyle", "gray")
		} else {
			drawCtx.Set("strokeStyle", "black")
		}

		c.transformUnscaled(drawCtx, mapping.Position.X(), mapping.Position.Y())
		if mapping.Suggested {
			drawCtx.Set("fillStyle", "rgba(255, 255, 255, 0.25)")
		} else {
			drawCtx.Set("fillStyle", "green")
		}
		drawCtx.Call("beginPath")
		drawCtx.Call("moveTo", 0, 0)
		drawCtx.Call("lineTo", 0, 0)
		drawCtx.Call("lineTo", 0, -20)
		drawCtx.Call("closePath")
		drawCtx.Set("shadowBlur", 0)
		drawCtx.Call("stroke")

		drawCtx.Call("rect", 0, -20, 15, 10)
		drawCtx.Call("fill")
		drawCtx.Call("stroke")

		drawCtx.Call("beginPath")
		drawCtx.Call("arc", 0, 0, 5, 0, 2*math.Pi, false)
		drawCtx.Set("shadowBlur", 5)
		drawCtx.Call("stroke")

		drawCtx.Set("fillStyle", "black")
		drawCtx.Set("font", "10px Arial")
		if pointOk {
			drawCtx.Call("fillText", point.Name, 8, 0)
		} else {
			drawCtx.Call("fillText", "Not mapped!", 8, 0)
		}

		if !mapping.Suggested {
			drawCtx.Call("beginPath")
			drawCtx.Call("moveTo", 0, 0)
			c.transformUnscaled(drawCtx, mapping.projectedPos.X(), mapping.projectedPos.Y())
			drawCtx.Call("lineTo", 0, 0)
			drawCtx.Call("stroke")
		}
	}

}
