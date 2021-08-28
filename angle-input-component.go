// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"strconv"
	"strings"

	"github.com/vugu/vugu"
)

type AngleInputComponent struct {
	BindValue *Angle
	BindLock  *bool
	LabelText string

	AttrMap vugu.AttrMap
}

func (c *AngleInputComponent) handleValueChange(event vugu.DOMEvent) {
	strVal := event.PropString("target", "value")
	strVal = strings.ReplaceAll(strVal, ",", ".")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return
	}

	c.BindValue.SetDegrees(val)
}

func (c *AngleInputComponent) handleLockChange(event vugu.DOMEvent) {
	*c.BindLock = event.PropBool("target", "checked")
}
