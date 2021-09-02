package main

import "github.com/vugu/vugu"

type ToggleInputComponent struct {
	AttrMap vugu.AttrMap

	BindValue *bool
	LabelText string
}

func (c *ToggleInputComponent) handleChange(event vugu.DOMEvent) {
	*c.BindValue = event.PropBool("target", "checked")
}
