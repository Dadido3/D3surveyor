// Copyright (C) 2021 David Vogel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"strconv"
	"strings"

	"github.com/vugu/vugu"
)

type AngleInputComponent struct {
	BindValue  *Angle
	BindLocked *bool
	LabelText  string

	AttrMap vugu.AttrMap
}

func (c *AngleInputComponent) handleValueChange(event vugu.DOMEvent) {
	strVal := event.PropString("target", "value")
	strVal = strings.ReplaceAll(strVal, ",", ".")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return
	}

	c.BindValue.SetDegree(val)
}

func (c *AngleInputComponent) handleLockedChange(event vugu.DOMEvent) {
	*c.BindLocked = event.PropBool("target", "checked")
}
