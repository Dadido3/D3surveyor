package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/vugu/vugu"
	js "github.com/vugu/vugu/js"
)

type CameraPhotoComponentEventCoordinate struct {
	xCan, yCan float64 // Position in canvas coordinates.
}

type CameraPhotoComponent struct {
	Photo *CameraPhoto

	// Canvas state variables. // TODO: Put most of the "scrollable/zoomable canvas" logic in its own module for reusability
	scale                     float64 // Ratio between Canvas and virtual coordinates.
	originX, originY          float64 // Origin in canvas coordinates. // TODO: Use specific type for DOM, Canvas and virtual coordinates, so that you can't mix them up
	canWidth, canHeight       float64 // Width and height in canvas pixels.
	canWidthDOM, canHeightDOM float64 // Width and height of the canvas in dom pixels.

	cachedImg js.Value // Cached js image object.

	ongoingMouseDrags map[int]CameraPhotoComponentEventCoordinate
	ongoingTouches    map[int]CameraPhotoComponentEventCoordinate

	selectedPoint    *CameraPhotoPoint
	highlightedPoint *CameraPhotoPoint
	//draggingPoint    *CameraPhotoPoint
}

func (c *CameraPhotoComponent) canvasCreated(canvas js.Value) {
	// TODO: Put this into a resize event or something similar
	c.canWidth, c.canHeight = canvas.Get("width").Float(), canvas.Get("height").Float()
	rect := canvas.Call("getBoundingClientRect")
	c.canWidthDOM, c.canHeightDOM = rect.Get("width").Float(), rect.Get("height").Float()

	c.canvasRedraw(canvas)
}

func (c *CameraPhotoComponent) Init(ctx vugu.InitCtx) {
	if c.ongoingMouseDrags == nil {
		c.ongoingMouseDrags = make(map[int]CameraPhotoComponentEventCoordinate)
	}
	if c.ongoingTouches == nil {
		c.ongoingTouches = make(map[int]CameraPhotoComponentEventCoordinate)
	}

	if c.scale == 0 {
		c.setScale(1, 0, 0)
	}
}

func (c *CameraPhotoComponent) handleUnmap(event vugu.DOMEvent) {
	for key, point := range c.Photo.points {
		if point == c.selectedPoint {
			c.selectedPoint = nil
			delete(c.Photo.points, key)
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
	xCan, yCan := c.transformDOMToCanvas(jsEvent.Get("offsetX").Float(), jsEvent.Get("offsetY").Float())
	switch inputType {
	case "mouse":
		// Store current pointer.
		c.ongoingMouseDrags[pointerID] = CameraPhotoComponentEventCoordinate{
			xCan: xCan,
			yCan: yCan,
		}
		// Check if current selection is still the one under the pointer.
		point, _, _ := c.getClosestPoint(xCan, yCan, 20*20)
		if c.selectedPoint != point {
			c.selectedPoint = nil
		}

	case "touch":
		// Store current pointer.
		c.ongoingTouches[pointerID] = CameraPhotoComponentEventCoordinate{
			xCan: xCan,
			yCan: yCan,
		}
		// Check if current selection is still the one under the pointer.
		point, _, _ := c.getClosestPoint(xCan, yCan, 20*20)
		if c.selectedPoint != point {
			c.selectedPoint = nil
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
	xCan, yCan := c.transformDOMToCanvas(jsEvent.Get("offsetX").Float(), jsEvent.Get("offsetY").Float())
	switch inputType {
	case "mouse":
		if ongoingMouseDrag, ok := c.ongoingMouseDrags[pointerID]; ok {
			if c.selectedPoint != nil {
				// Drag selected point.
				c.selectedPoint.x += (xCan - ongoingMouseDrag.xCan) / c.scale / float64(c.Photo.ImageConf.Width)
				c.selectedPoint.y += (yCan - ongoingMouseDrag.yCan) / c.scale / float64(c.Photo.ImageConf.Height)
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
		point, _, _ := c.getClosestPoint(xCan, yCan, 20*20)
		if c.highlightedPoint != point {
			c.highlightedPoint = point
		}

	case "touch":
		switch len(c.ongoingTouches) {
		case 1:
			if ongoingTouch, ok := c.ongoingTouches[pointerID]; ok {
				if c.selectedPoint != nil {
					// Drag selected point.
					c.selectedPoint.x += (xCan - ongoingTouch.xCan) / c.scale / float64(c.Photo.ImageConf.Width)
					c.selectedPoint.y += (yCan - ongoingTouch.yCan) / c.scale / float64(c.Photo.ImageConf.Height)
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
						prevDistance := math.Sqrt(math.Pow(ongoingTouch.xCan-ongoingOtherTouch.xCan, 2) + math.Pow(ongoingTouch.yCan-ongoingOtherTouch.yCan, 2))
						newDistance := math.Sqrt(math.Pow(xCan-ongoingOtherTouch.xCan, 2) + math.Pow(yCan-ongoingOtherTouch.yCan, 2))
						pivotX, pivotY := c.transformCanvasToVirtual((ongoingTouch.xCan+ongoingOtherTouch.xCan)/2, (ongoingTouch.yCan+ongoingOtherTouch.yCan)/2)
						c.setScale(c.scale*(newDistance/prevDistance), pivotX, pivotY)
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
	xCan, yCan := c.transformDOMToCanvas(jsEvent.Get("offsetX").Float(), jsEvent.Get("offsetY").Float())
	switch inputType {
	case "mouse":
		jsCanvas.Call("releasePointerCapture", pointerID)
		// Reset highlighted point.
		c.highlightedPoint = nil
		// Get element selection.
		point, _, _ := c.getClosestPoint(xCan, yCan, 20*20)
		if c.selectedPoint != point {
			c.selectedPoint = point
		}
		delete(c.ongoingMouseDrags, pointerID)

	case "touch":
		jsCanvas.Call("releasePointerCapture", pointerID)

		// Get element selection.
		point, _, _ := c.getClosestPoint(xCan, yCan, 20*20)
		if c.selectedPoint != point {
			c.selectedPoint = point
		}

		delete(c.ongoingTouches, pointerID)

	default:
		fmt.Printf("Input type %q not supported\n", inputType)
	}
}

func (c *CameraPhotoComponent) handleDblClick(event vugu.DOMEvent) {
	jsEvent, _ := event.JSEvent(), event.JSEventTarget()

	xCan, yCan := c.transformDOMToCanvas(jsEvent.Get("offsetX").Float(), jsEvent.Get("offsetY").Float())
	xVir, yVir := c.transformCanvasToVirtual(xCan, yCan)

	point, _, _ := c.getClosestPoint(xCan, yCan, 20*20)
	if point == nil {
		point := c.Photo.NewPoint()
		point.x, point.y = xVir/float64(c.Photo.ImageConf.Width), yVir/float64(c.Photo.ImageConf.Height)
	}
}

func (c *CameraPhotoComponent) handleClick(event vugu.DOMEvent) {

}

func (c *CameraPhotoComponent) handleWheel(event vugu.DOMEvent) {
	jsEvent, _ := event.JSEvent(), event.JSEventTarget()

	jsEvent.Call("preventDefault")
	jsEvent.Call("stopPropagation")

	delta := jsEvent.Get("deltaY").Float()
	xPivotCan, yPivotCan := c.transformDOMToCanvas(jsEvent.Get("offsetX").Float(), jsEvent.Get("offsetY").Float())
	xPivot, yPivot := c.transformCanvasToVirtual(xPivotCan, yPivotCan)

	c.setScale(math.Pow(1/1.001, delta)*c.scale, xPivot, yPivot)
}

// transformDOMToCanvas takes the coordinates relative to the top left of the element in DOM pixels and transforms them into the canvas coordinates.
func (c *CameraPhotoComponent) transformDOMToCanvas(xDOM, yDOM float64) (xCan, yCan float64) {
	return xDOM / c.canWidthDOM * c.canWidth, yDOM / c.canHeightDOM * c.canHeight
}

// transformCanvasToVirtual takes canvas coordinates and transforms them into virtual coordinates.
func (c *CameraPhotoComponent) transformCanvasToVirtual(xCan, yCan float64) (xVir, yVir float64) {
	return (xCan - c.originX) / c.scale, (yCan - c.originY) / c.scale
}

// transformVirtualToCanvas takes virtual coordinates and transforms them into canvas coordinates.
func (c *CameraPhotoComponent) transformVirtualToCanvas(xVir, yVir float64) (xCan, yCan float64) {
	return xVir*c.scale + c.originX, yVir*c.scale + c.originY
}

// setScale clamps and overwrites the current scale.
// xPivot and yPivot are in virtual coordinates.
func (c *CameraPhotoComponent) setScale(newScale, xPivot, yPivot float64) {
	// Clamp new scale.
	if newScale > 10 {
		newScale = 10
	} else if newScale < 0.01 {
		newScale = 0.01
	}

	// Calculate new origin in canvas coordinates.
	c.originX, c.originY = c.originX+xPivot*(c.scale-newScale), c.originY+yPivot*(c.scale-newScale)

	c.scale = newScale
}

// transformUnscaled sets the transformation matrix to have its origin at the given virtual coordinate, but the scale is set to 1.
func (c *CameraPhotoComponent) transformUnscaled(drawCtx js.Value, xVir, yVir float64) {
	xCan, yCan := c.transformVirtualToCanvas(xVir, yVir)
	drawCtx.Call("setTransform", 1, 0, 0, 1, xCan, yCan)
}

// transformScaled sets the transformation matrix to represent the origin and scaling values.
func (c *CameraPhotoComponent) transformScaled(drawCtx js.Value) {
	drawCtx.Call("setTransform", c.scale, 0, 0, c.scale, c.originX, c.originY)
}

// getClosestPoint returns the closest mapped point to the given canvas coordinates.
func (c *CameraPhotoComponent) getClosestPoint(xCan, yCan, maxDistSqr float64) (minPoint *CameraPhotoPoint, minKey string, minDistSqr float64) {
	minDistSqr = maxDistSqr

	for key, point := range c.Photo.points {
		pXCan, pYCan := c.transformVirtualToCanvas(point.x*float64(c.Photo.ImageConf.Width), point.y*float64(c.Photo.ImageConf.Height))
		distSqr := math.Pow(pXCan-xCan, 2) + math.Pow(pYCan-yCan, 2)
		if minDistSqr > distSqr {
			minDistSqr, minKey, minPoint = distSqr, key, point
		}
	}

	return
}

func (c *CameraPhotoComponent) canvasRedraw(canvas js.Value) {
	drawCtx := canvas.Call("getContext", "2d")

	if c.cachedImg.IsUndefined() {
		c.cachedImg = js.Global().Get("Image").New()
		c.cachedImg.Set("src", c.Photo.jsImageURL)
	}

	drawCtx.Set("shadowBlur", 0)

	drawCtx.Call("setTransform", 1, 0, 0, 1, 0, 0)
	drawCtx.Call("clearRect", 0, 0, c.canWidth, c.canHeight)
	c.transformScaled(drawCtx)

	drawCtx.Call("drawImage", c.cachedImg, 0, 0)

	c.transformUnscaled(drawCtx, 10, 50)
	drawCtx.Set("fillStyle", "white")
	drawCtx.Set("font", "30px Arial")
	drawCtx.Call("fillText", fmt.Sprintf("image %d, %d, %f", c.Photo.ImageConf.Width, c.Photo.ImageConf.Height, rand.Float64()), 0, 0)

	drawCtx.Set("lineWidth", 1)
	drawCtx.Set("lineCap", "butt")
	drawCtx.Set("fillStyle", "green")
	drawCtx.Set("shadowOffsetX", 0)
	drawCtx.Set("shadowOffsetY", 0)
	drawCtx.Set("shadowColor", "white")
	for _, point := range c.Photo.points {
		if point == c.selectedPoint {
			drawCtx.Set("strokeStyle", "white")
		} else if point == c.highlightedPoint {
			drawCtx.Set("strokeStyle", "gray")
		} else {
			drawCtx.Set("strokeStyle", "black")
		}

		c.transformUnscaled(drawCtx, point.x*float64(c.Photo.ImageConf.Width), point.y*float64(c.Photo.ImageConf.Height))
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

	}

}
