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

// Coordinate represents a point in the world.
// This may also be a relative to another point.
type Coordinate [3]Distance

func (c Coordinate) X() Distance {
	return c[0]
}

func (c Coordinate) Y() Distance {
	return c[1]
}

func (c Coordinate) Z() Distance {
	return c[2]
}

// Distance returns the distance between itself and the second coordinate.
func (c Coordinate) Distance(c2 Coordinate) Distance {
	x, y, z := c[0]-c2[0], c[1]-c2[1], c[2]-c2[2]
	sqrSum := float64(x*x + y*y + z*z)
	return Distance(math.Sqrt(sqrSum))
}

func (c Coordinate) Add(c2 Coordinate) Coordinate {
	return Coordinate{c[0] + c2[0], c[1] + c2[1], c[2] + c2[2]}
}

func (c Coordinate) Sub(c2 Coordinate) Coordinate {
	return Coordinate{c[0] - c2[0], c[1] - c2[1], c[2] - c2[2]}
}

// Vec3 returns the coordinate as vector.
// Its unit is in meters.
func (c Coordinate) Vec3() mgl64.Vec3 {
	return mgl64.Vec3{float64(c[0]), float64(c[1]), float64(c[2])}
}

// Vec4 returns the coordinate as vector.
// Its unit is in meters.
func (c Coordinate) Vec4(w float64) mgl64.Vec4 {
	return mgl64.Vec4{float64(c[0]), float64(c[1]), float64(c[2]), w}
}
