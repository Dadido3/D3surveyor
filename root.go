package main

import (
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
