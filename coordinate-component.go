package main

import (
	"strconv"
	"strings"

	"github.com/vugu/vugu"
)

type CoordinateComponent struct {
	BindValue *Coordinate

	Editable bool
}

func (c *CoordinateComponent) handleChange(event vugu.DOMEvent) {
	strVal := event.PropString("target", "value")
	strVal = strings.ReplaceAll(strVal, ",", ".")

	name := event.PropString("target", "name")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return
	}

	switch name {
	case "X":
		c.BindValue.X = val
	case "Y":
		c.BindValue.Y = val
	case "Z":
		c.BindValue.Z = val
	}
}
