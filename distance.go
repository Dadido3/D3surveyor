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
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Distance describes a distance in meters, or an absolute position measured by its distance from the origin.
type Distance float64

func (d Distance) Meters() float64 {
	return float64(d)
}

// TweakableValue returns the values mapped into optimizer space.
func (d Distance) TweakableValue() float64 {
	return float64(d)
}

// SetTweakableValue converts and applies the given value from optimizer space.
func (d *Distance) SetTweakableValue(v float64) {
	*d = Distance(v)
}

// InputValue implements the valuer interface of the general input component.
func (d Distance) InputValue() string {
	return fmt.Sprintf("%.13g", d)
}

// SetInputValue implements the valuer interface of the general input component.
func (d *Distance) SetInputValue(strVal string) {
	strVal = strings.ReplaceAll(strVal, ",", ".")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		log.Printf("strconv.ParseFloat() failed: %v", err)
		return
	}

	*d = Distance(val)
}

func (d Distance) Sqr() float64 {
	return float64(d * d)
}
