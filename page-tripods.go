// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import "github.com/vugu/vgrouter"

type PageTripods struct {
	vgrouter.NavigatorRef `json:"-"`

	Site *Site
}

func (c *PageTripods) handleAdd() {
	tripod := c.Site.NewTripod("")

	c.Navigate("/tripod/"+tripod.Key(), nil)
}
