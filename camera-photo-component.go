package main

import (
	"fmt"
	"math"
	"math/rand"

	js "github.com/vugu/vugu/js"
)

type CameraPhotoComponentEventCoordinate struct {
	xCan, yCan float64 // Position in canvas coordinates.
}

type CameraPhotoComponent struct {
	Photo *CameraPhoto

	drawCtx js.Value
	img     js.Value // A JS image object containing the photo via URL to the blob. The data is stored in the CameraPhoto object "Photo".

	// Canvas state variables. // TODO: Put most of the "scrollable/zoomable canvas" logic in its own module for reusability
	scale               float64 // Ratio between Canvas and virtual coordinates.
	originX, originY    float64 // Origin in canvas coordinates. // TODO: Use specific type for DOM, Canvas and virtual coordinates, so that you can't mix them up
	canWidth, canHeight float64 // Width and height in canvas pixels.

	canWidthDOM, canHeightDOM float64 // Width and height of the canvas in dom pixels.

	ongoingMouseDrags map[int]CameraPhotoComponentEventCoordinate
	ongoingTouches    map[int]CameraPhotoComponentEventCoordinate

	selectedPoint    *CameraPhotoPoint
	highlightedPoint *CameraPhotoPoint
	//draggingPoint    *CameraPhotoPoint
}

func (c *CameraPhotoComponent) canvasCreated(canvas js.Value) {

	c.canWidth, c.canHeight = canvas.Get("width").Float(), canvas.Get("height").Float()
	rect := canvas.Call("getBoundingClientRect")
	c.canWidthDOM, c.canHeightDOM = rect.Get("width").Float(), rect.Get("height").Float()

	c.drawCtx = canvas.Call("getContext", "2d")

	c.img = js.Global().Get("Image").New()
	c.img.Set("src", c.Photo.jsImageURL)

	if c.scale == 0 {
		c.setScale(1, 0, 0)
	}

	if c.ongoingMouseDrags == nil {
		c.ongoingMouseDrags = make(map[int]CameraPhotoComponentEventCoordinate)
	}
	if c.ongoingTouches == nil {
		c.ongoingTouches = make(map[int]CameraPhotoComponentEventCoordinate)
	}

	/*canvas.Call("addEventListener", "contextmenu", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		//event.Call("preventDefault")
		//event.Call("stopPropagation")

		return js.Undefined()
	}))*/

	canvas.Call("addEventListener", "pointerdown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		//event.Call("preventDefault")
		//event.Call("stopPropagation")

		pointerID := event.Get("pointerId").Int()
		inputType := event.Get("pointerType").String()
		switch inputType {
		case "mouse":
			canvas.Call("setPointerCapture", pointerID)
			xCan, yCan := c.transformDOMToCanvas(event.Get("offsetX").Float(), event.Get("offsetY").Float())
			c.ongoingMouseDrags[pointerID] = CameraPhotoComponentEventCoordinate{
				xCan: xCan,
				yCan: yCan,
			}

		case "touch":
			canvas.Call("setPointerCapture", event.Get("pointerId"))
			xCan, yCan := c.transformDOMToCanvas(event.Get("offsetX").Float(), event.Get("offsetY").Float())
			c.ongoingTouches[pointerID] = CameraPhotoComponentEventCoordinate{
				xCan: xCan,
				yCan: yCan,
			}

		default:
			fmt.Printf("Input type %q not supported\n", inputType)
		}

		return js.Undefined()
	}))

	canvas.Call("addEventListener", "pointermove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		//event.Call("preventDefault")
		//event.Call("stopPropagation")

		pointerID := event.Get("pointerId").Int()
		inputType := event.Get("pointerType").String()
		switch inputType {
		case "mouse":
			xCan, yCan := c.transformDOMToCanvas(event.Get("offsetX").Float(), event.Get("offsetY").Float())
			if ongoingMouseDrag, ok := c.ongoingMouseDrags[pointerID]; ok {
				c.originX += xCan - ongoingMouseDrag.xCan
				c.originY += yCan - ongoingMouseDrag.yCan

				c.ongoingMouseDrags[pointerID] = CameraPhotoComponentEventCoordinate{
					xCan: xCan,
					yCan: yCan,
				}
				c.canvasRedraw(canvas)
			}

			// Get highlighted element.
			point, _, _ := c.getClosestPoint(xCan, yCan, 20*20)
			if c.highlightedPoint != point {
				c.highlightedPoint = point
				c.canvasRedraw(canvas)
			}

		case "touch":
			switch len(c.ongoingTouches) {
			case 1:
				if ongoingTouch, ok := c.ongoingTouches[pointerID]; ok {
					xCan, yCan := c.transformDOMToCanvas(event.Get("offsetX").Float(), event.Get("offsetY").Float())
					c.originX += xCan - ongoingTouch.xCan
					c.originY += yCan - ongoingTouch.yCan

					c.ongoingTouches[pointerID] = CameraPhotoComponentEventCoordinate{
						xCan: xCan,
						yCan: yCan,
					}
				}
			case 2:
				if ongoingTouch, ok := c.ongoingTouches[pointerID]; ok {
					xCan, yCan := c.transformDOMToCanvas(event.Get("offsetX").Float(), event.Get("offsetY").Float())

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

			c.canvasRedraw(canvas)

		default:
			fmt.Printf("Input type %q not supported\n", inputType)
		}

		return js.Undefined()
	}))

	handlePointerUp := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		//event.Call("preventDefault")
		//event.Call("stopPropagation")

		pointerID := event.Get("pointerId").Int()
		inputType := event.Get("pointerType").String()
		switch inputType {
		case "mouse":
			canvas.Call("releasePointerCapture", pointerID)
			c.highlightedPoint = nil
			delete(c.ongoingMouseDrags, pointerID)

		case "touch":
			canvas.Call("releasePointerCapture", pointerID)
			delete(c.ongoingTouches, pointerID)

		default:
			fmt.Printf("Input type %q not supported\n", inputType)
		}

		return js.Undefined()
	})
	canvas.Call("addEventListener", "pointerup", handlePointerUp)
	canvas.Call("addEventListener", "pointercancel", handlePointerUp)
	canvas.Call("addEventListener", "pointerout", handlePointerUp)
	canvas.Call("addEventListener", "pointerLeave", handlePointerUp)

	canvas.Call("addEventListener", "dblclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]

		xCan, yCan := c.transformDOMToCanvas(event.Get("offsetX").Float(), event.Get("offsetY").Float())
		xVir, yVir := c.transformCanvasToVirtual(xCan, yCan)

		point := c.Photo.NewPoint()
		point.x, point.y = xVir/float64(c.Photo.ImageConf.Width), yVir/float64(c.Photo.ImageConf.Height)

		c.canvasRedraw(canvas)

		return js.Undefined()
	}))

	canvas.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]

		xCan, yCan := c.transformDOMToCanvas(event.Get("offsetX").Float(), event.Get("offsetY").Float())

		// Get element selection.
		point, _, _ := c.getClosestPoint(xCan, yCan, 20*20)
		if c.selectedPoint != point {
			c.selectedPoint = point
			c.canvasRedraw(canvas)
		}

		return js.Undefined()
	}))

	/*canvas.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		//event.Call("preventDefault")
		//event.Call("stopPropagation")

		if event.Get("button").Int() == 0 {
			c.mouseDragging = true
			canvas.Call("setPointerCapture", event.Get("pointerId"))
		}

		return js.Undefined()
	}))

	canvas.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		//event.Call("preventDefault")
		//event.Call("stopPropagation")

		if c.mouseDragging {
			c.originX += event.Get("movementX").Float()
			c.originY += event.Get("movementY").Float()
			c.canvasRedraw(canvas)
		}

		return js.Undefined()
	}))

	canvas.Call("addEventListener", "mouseup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		//event.Call("preventDefault")
		//event.Call("stopPropagation")

		if event.Get("button").Int() == 0 {
			c.mouseDragging = false
			canvas.Call("releasePointerCapture", event.Get("pointerId"))
		}

		return js.Undefined()
	}))*/

	canvas.Call("addEventListener", "wheel", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")
		event.Call("stopPropagation")

		delta := event.Get("deltaY").Float()
		xPivotCan, yPivotCan := c.transformDOMToCanvas(event.Get("offsetX").Float(), event.Get("offsetY").Float())
		xPivot, yPivot := c.transformCanvasToVirtual(xPivotCan, yPivotCan)

		c.setScale(math.Pow(1/1.001, delta)*c.scale, xPivot, yPivot)
		c.canvasRedraw(canvas)

		return js.Undefined()
	}))

	c.canvasRedraw(canvas)
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
func (c *CameraPhotoComponent) transformUnscaled(xVir, yVir float64) {
	xCan, yCan := c.transformVirtualToCanvas(xVir, yVir)
	c.drawCtx.Call("setTransform", 1, 0, 0, 1, xCan, yCan)
}

// transformScaled sets the transformation matrix to represent the origin and scaling values.
func (c *CameraPhotoComponent) transformScaled() {
	c.drawCtx.Call("setTransform", c.scale, 0, 0, c.scale, c.originX, c.originY)
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
	c.drawCtx.Set("shadowBlur", 0)

	c.drawCtx.Call("setTransform", 1, 0, 0, 1, 0, 0)
	c.drawCtx.Call("clearRect", 0, 0, c.canWidth, c.canHeight)
	c.transformScaled()

	c.drawCtx.Call("drawImage", c.img, 0, 0)

	c.transformUnscaled(10, 50)
	c.drawCtx.Set("fillStyle", "white")
	c.drawCtx.Set("font", "30px Arial")
	c.drawCtx.Call("fillText", fmt.Sprintf("image %d, %d, %f", c.Photo.ImageConf.Width, c.Photo.ImageConf.Height, rand.Float64()), 0, 0)

	c.drawCtx.Set("lineWidth", 1)
	c.drawCtx.Set("lineCap", "butt")
	c.drawCtx.Set("fillStyle", "green")
	c.drawCtx.Set("shadowOffsetX", 0)
	c.drawCtx.Set("shadowOffsetY", 0)
	c.drawCtx.Set("shadowColor", "white")
	for _, point := range c.Photo.points {
		if point == c.selectedPoint {
			c.drawCtx.Set("strokeStyle", "white")
		} else if point == c.highlightedPoint {
			c.drawCtx.Set("strokeStyle", "gray")
		} else {
			c.drawCtx.Set("strokeStyle", "black")
		}

		c.transformUnscaled(point.x*float64(c.Photo.ImageConf.Width), point.y*float64(c.Photo.ImageConf.Height))
		c.drawCtx.Call("beginPath")
		c.drawCtx.Call("moveTo", 0, 0)
		c.drawCtx.Call("lineTo", 0, 0)
		c.drawCtx.Call("lineTo", 0, -20)
		c.drawCtx.Call("closePath")
		c.drawCtx.Set("shadowBlur", 0)
		c.drawCtx.Call("stroke")

		c.drawCtx.Call("rect", 0, -20, 15, 10)
		c.drawCtx.Call("fill")
		c.drawCtx.Call("stroke")

		c.drawCtx.Call("beginPath")
		c.drawCtx.Call("arc", 0, 0, 5, 0, 2*math.Pi, false)
		c.drawCtx.Set("shadowBlur", 5)
		c.drawCtx.Call("stroke")

	}

}
