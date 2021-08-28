// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import "github.com/vugu/vgrouter"

type PageRangefinders struct {
	vgrouter.NavigatorRef `json:"-"`

	Site *Site
}

func (c *PageRangefinders) handleAdd() {
	rangefinder := c.Site.NewRangefinder("")

	c.Navigate("/rangefinder/"+rangefinder.Key(), nil)
}
