package main

import (
	"strconv"
	"strings"

	"github.com/vugu/vugu"
)

type DistanceInputComponent struct {
	BindValue *float64

	AttrMap vugu.AttrMap
}

func (c *DistanceInputComponent) handleChange(event vugu.DOMEvent) {
	strVal := event.PropString("target", "value")
	strVal = strings.ReplaceAll(strVal, ",", ".")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return
	}

	*c.BindValue = val

}
