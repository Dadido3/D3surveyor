package main

import (
	"github.com/vugu/vugu"
	"github.com/vugu/vugu/vgform"
)

type PointSelectionComponent struct {
	Site *Site

	BindValue *string

	AttrMap vugu.AttrMap

	options vgform.MapOptions
}

func (c *PointSelectionComponent) Init(ctx vugu.InitCtx) {

	// Only load the list at creation of the component.
	// Updating it would cause the dropdown to show the wrong option.
	c.options = vgform.MapOptions{"": ""}

	for _, point := range c.Site.Points {
		c.options[point.Key()] = point.Name
	}
}
