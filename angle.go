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

import "math"

// Angle describes a angle in radian.
type Angle float64

// TweakableValue returns the values mapped into optimizer space.
func (a *Angle) TweakableValue() float64 {
	return float64(*a)
}

// SetTweakableValue converts and applies the given value from optimizer space.
func (a *Angle) SetTweakableValue(v float64) {
	*a = Angle(v).Normalized()
}

// Normalized returns the angle in the range of [0,2π).
func (a Angle) Normalized() Angle {
	rad := math.Remainder(float64(a), 2*math.Pi)
	if rad < 0 {
		rad += 2 * math.Pi
	}

	return Angle(rad)
}

func (a Angle) Degrees() float64 {
	return float64(a) * (180 / math.Pi)
}

func (a *Angle) SetDegrees(deg float64) {
	*a = Angle(deg * (math.Pi / 180))
}
