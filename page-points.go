// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import "github.com/vugu/vgrouter"

type PagePoints struct {
	vgrouter.NavigatorRef `json:"-"`

	Site *Site
}

func (c *PagePoints) handleAdd() {
	p := c.Site.NewPoint("")

	c.Navigate("/point/"+p.Key(), nil)
}
