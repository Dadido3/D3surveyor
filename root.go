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

	Body vugu.Builder
}

func (r *Root) handleRecalc(event vugu.DOMEvent) {
	Optimize(globalSite)
}

func (r *Root) handleDownload(event vugu.DOMEvent) {
	data, err := json.Marshal(globalSite)
	if err != nil {
		log.Printf("json.Marshal failed: %v", err)
	}
	browserDownload(fmt.Sprintf("%v.D3mula", globalSite.Name), data, "text/json")
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
			// TODO: Somehow tell the user the image couldn't be loaded
			return js.Undefined()
		}

		globalSite = newSite

		r.Navigate("/", nil)

		return js.Undefined()
	}))

	imgFile := js.Global().Get("document").Call("getElementById", "site-upload").Get("files").Index(0)
	fileReader.Call("readAsArrayBuffer", imgFile)
}
