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
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

type Coordinate struct {
	X, Y, Z             Distance
	LockX, LockY, LockZ bool // Lock (Don't optimize) the value.
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (c *Coordinate) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables := make([]Tweakable, 0, 3)
	if !c.LockX {
		tweakables = append(tweakables, &c.X)
	}
	if !c.LockY {
		tweakables = append(tweakables, &c.Y)
	}
	if !c.LockZ {
		tweakables = append(tweakables, &c.Z)
	}

	return tweakables, nil
}

// Distance returns the distance between itself and the second coordinate.
func (c *Coordinate) Distance(c2 Coordinate) Distance {
	sqr := (c.X-c2.X)*(c.X-c2.X) + (c.Y-c2.Y)*(c.Y-c2.Y) + (c.Z-c2.Z)*(c.Z-c2.Z)
	return Distance(math.Sqrt(float64(sqr)))
}

// Vec3 returns the coordinate as vector.
// Its unit is in meters.
func (c Coordinate) Vec3() mgl64.Vec3 {
	return mgl64.Vec3{float64(c.X), float64(c.Y), float64(c.Z)}
}
