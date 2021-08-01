package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
)

type Root struct {
	vgrouter.NavigatorRef

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
