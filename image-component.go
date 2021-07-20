package main

import (
	"github.com/vugu/vugu/js"
)

type ImageComponent struct {
	BindValue *[]byte
}

func (c *ImageComponent) handlePopulate(imgElement js.Value) {
	if c.BindValue == nil {
		return
	}

	dst := js.Global().Get("Uint8Array").New(len(*c.BindValue))
	js.CopyBytesToJS(dst, *c.BindValue)

	dstArray := js.Global().Get("Array").New(dst)
	blob := js.Global().Get("Blob").New(dstArray, js.ValueOf(map[string]interface{}{"type": "image/*"}))
	url := js.Global().Get("URL").Call("createObjectURL", blob)
	imgElement.Set("src", url)

}
