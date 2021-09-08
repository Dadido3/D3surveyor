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

// PixelCoordinate represents a position on a photo measured in pixels.
// Origin is at the top left.
// The Z component ranges from 0 being the at far end clipping plane to 1 being at the near end clipping plane. // TODO: Check Z direction for correctness, it may be swapped
type PixelCoordinate [3]PixelDistance

func (p PixelCoordinate) X() PixelDistance {
	return p[0]
}

func (p PixelCoordinate) Y() PixelDistance {
	return p[1]
}

func (p PixelCoordinate) Z() PixelDistance {
	return p[2]
}

// IsZero returns whether all components are 0.
func (p PixelCoordinate) IsZero() bool {
	if p[0] == 0 && p[1] == 0 && p[2] == 0 {
		return true
	}
	return false
}

// LengthSqr returns the distance from the origin in pixels squared.
// This ignores the Z component, as it's not measured in pixels.
func (p PixelCoordinate) LengthSqr() float64 {
	x, y := p[0], p[1]
	return float64(x*x + y*y)
}

// Length returns the distance from the origin.
// This ignores the Z component, as it's not measured in pixels.
func (p PixelCoordinate) Length() PixelDistance {
	return PixelDistance(math.Sqrt(p.LengthSqr()))
}

// DistanceSqr returns the squared distance between itself and the second coordinate in pixels squared.
// This ignores the Z component, as it's not measured in pixels.
func (p PixelCoordinate) DistanceSqr(p2 PixelCoordinate) float64 {
	return p.Sub(p2).LengthSqr()
}

// Distance returns the distance between itself and the second coordinate.
// This ignores the Z component, as it's not measured in pixels.
func (p PixelCoordinate) Distance(p2 PixelCoordinate) PixelDistance {
	return p.Sub(p2).Length()
}

// Scaled returns the pixel coordinate scaled by the factor s around its origin.
// This ignores the Z component, as it's not measured in pixels.
func (p PixelCoordinate) Scaled(s float64) PixelCoordinate {
	p[0] *= PixelDistance(s)
	p[1] *= PixelDistance(s)
	return p
}

func (p PixelCoordinate) Add(p2 PixelCoordinate) PixelCoordinate {
	return PixelCoordinate{p[0] + p2[0], p[1] + p2[1], p[2] + p2[2]}
}

func (p PixelCoordinate) Sub(p2 PixelCoordinate) PixelCoordinate {
	return PixelCoordinate{p[0] - p2[0], p[1] - p2[1], p[2] - p2[2]}
}
