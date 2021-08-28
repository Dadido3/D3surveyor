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

func (r *Root) handleRecalculate(event vugu.DOMEvent) {
	go Optimize(event.EventEnv(), globalSite)
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
