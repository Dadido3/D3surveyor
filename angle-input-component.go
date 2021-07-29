package main

import (
	"strconv"
	"strings"

	"github.com/vugu/vugu"
)

type AngleInputComponent struct {
	BindValue *Angle

	AttrMap vugu.AttrMap
}

func (c *AngleInputComponent) handleChange(event vugu.DOMEvent) {
	strVal := event.PropString("target", "value")
	strVal = strings.ReplaceAll(strVal, ",", ".")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return
	}

	*c.BindValue = Angle(val)

}
