// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import js "github.com/vugu/vugu/js"

func browserDownload(filename string, data []byte, mime string) {
	// Create js blob and URL.
	dst := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(dst, data)
	dstArray := js.Global().Get("Array").New(dst)
	blob := js.Global().Get("Blob").New(dstArray, js.ValueOf(map[string]interface{}{"type": mime}))

	url := js.Global().Get("URL").Call("createObjectURL", blob)
	defer js.Global().Get("URL").Call("revokeObjectURL", url)

	elem := js.Global().Get("document").Call("createElement", "a")
	elem.Set("href", url)
	elem.Set("download", filename)

	js.Global().Get("document").Get("body").Call("appendChild", elem)
	elem.Call("click")
	js.Global().Get("document").Get("body").Call("removeChild", elem)
}
