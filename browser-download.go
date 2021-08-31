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
