package main

import "github.com/vugu/vgrouter"

type PageCameras struct {
	vgrouter.NavigatorRef

	Site *Site
}

func (c *PageCameras) handleAdd() {
	camera := c.Site.NewCamera("adsf")

	c.Navigate("/camera/"+camera.Key(), nil)
}
