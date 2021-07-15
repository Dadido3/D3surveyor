package main

import "github.com/vugu/vgrouter"

type PagePoints struct {
	vgrouter.NavigatorRef

	Site *Site
}

func (c *PagePoints) handleAdd() {
	p := c.Site.NewPoint("adsf")

	c.Navigate("/point/"+p.Key(), nil)
}
