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
	"log"
	"strconv"
	"strings"

	"github.com/vugu/vugu"
)

type DistanceInputComponent struct {
	BindValue *Distance
	BindLock  *bool
	LabelText string

	AttrMap vugu.AttrMap
}

func (c *DistanceInputComponent) handleValueChange(event vugu.DOMEvent) {
	strVal := event.PropString("target", "value")
	strVal = strings.ReplaceAll(strVal, ",", ".")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		log.Printf("strconv.ParseFloat() failed: %v", err)
		return
	}

	*c.BindValue = Distance(val)
}

func (c *DistanceInputComponent) handleLockChange(event vugu.DOMEvent) {
	*c.BindLock = event.PropBool("target", "checked")
}
