package main

import "github.com/vugu/vgrouter"

type PageRangefinders struct {
	vgrouter.NavigatorRef

	Site *Site
}

func (c *PageRangefinders) handleAdd() {
	rangefinder := c.Site.NewRangefinder("adsf")

	c.Navigate("/rangefinder/"+rangefinder.Key(), nil)
}
