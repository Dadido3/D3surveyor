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

// PixelCoordinate represents a position on a photo measured in pixels.
// Origin is at the top left.
// The Z direction ranges from 0 being the at far end clipping plane to 1 being at the near end clipping plane. // TODO: Check Z direction for correctness, it may be swapped
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

func (p PixelCoordinate) Add(p2 PixelCoordinate) PixelCoordinate {
	return PixelCoordinate{p[0] + p2[0], p[1] + p2[1], p[2] + p2[2]}
}

func (p PixelCoordinate) Sub(p2 PixelCoordinate) PixelCoordinate {
	return PixelCoordinate{p[0] - p2[0], p[1] - p2[1], p[2] - p2[2]}
}
