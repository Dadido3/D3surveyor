// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"github.com/vugu/vgrouter"
)

type PageCameras struct {
	vgrouter.NavigatorRef `json:"-"`

	Site *Site
}

func (c *PageCameras) handleAdd() {
	camera := c.Site.NewCamera("")

	c.Navigate("/camera/"+camera.Key(), nil)
}
