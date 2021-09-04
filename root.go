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
	"encoding/json"
	"fmt"
	"log"

	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
	js "github.com/vugu/vugu/js"
)

type Root struct {
	vgrouter.NavigatorRef `json:"-"`

	sidebarDisplay string

	Body vugu.Builder
}

func (r *Root) handleSidebarOpen(event vugu.DOMEvent) {
	r.sidebarDisplay = "block"
}

func (r *Root) handleSidebarClose(event vugu.DOMEvent) {
	r.sidebarDisplay = "none"
}

func (r *Root) handleDownload(event vugu.DOMEvent) {
	data, err := json.MarshalIndent(globalSite, "", "\t")
	if err != nil {
		log.Printf("json.Marshal failed: %v", err)
	}

	browserDownload(fmt.Sprintf("%v.D3survey", globalSite.Name), data, "application/octet-stream")
}

func (r *Root) handleUploadClick(event vugu.DOMEvent) {
	js.Global().Get("document").Call("getElementById", "site-upload").Call("click")
}

func (r *Root) handleUpload(event vugu.DOMEvent) {
	fileReader := js.Global().Get("FileReader").New()
	fileReader.Call("addEventListener", "loadend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		buffer := fileReader.Get("result")
		uint8Array := js.Global().Get("Uint8Array").New(buffer)

		jsonData := make([]byte, uint8Array.Length())
		js.CopyBytesToGo(jsonData, uint8Array)

		event.EventEnv().Lock()
		defer event.EventEnv().UnlockRender()

		newSite, err := NewSiteFromJSON(jsonData)
		if err != nil {
			log.Printf("NewSiteFromJSON failed: %v", err)
			// TODO: Somehow tell the user the file couldn't be loaded
			return js.Undefined()
		}

		globalSite = newSite

		r.Navigate("/", nil)

		return js.Undefined()
	}))

	imgFiles := js.Global().Get("document").Call("getElementById", "site-upload").Get("files")
	if imgFiles.Length() != 1 {
		log.Printf("Wrong amount of files: Expected %v, got %v", 1, imgFiles.Length())
		// TODO: Somehow forward the error to the user
		return
	}
	fileReader.Call("readAsArrayBuffer", imgFiles.Index(0))
}

func (r *Root) handleExport(event vugu.DOMEvent) {
	data := generateObj(globalSite)

	browserDownload(fmt.Sprintf("%v.obj", globalSite.Name), data, "application/octet-stream")
}
