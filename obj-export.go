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

import "fmt"

// generateObj returns the features of the given site as a Wavefront OBJ file.
func generateObj(site *Site) []byte {
	result := "#List of points\n"
	pointKeyIndices := map[string]int{}
	counter := 1
	for key, point := range site.Points {
		pointKeyIndices[key] = counter
		counter++
		result += fmt.Sprintf("v %f %f %f\n", point.Position.X, point.Position.Y, point.Position.Z)
	}

	result += "\n#List of rangefinder measurements\n"
	for _, rangefinder := range site.Rangefinders {
		for _, measurement := range rangefinder.Measurements {
			indexP1, ok1 := pointKeyIndices[measurement.P1]
			indexP2, ok2 := pointKeyIndices[measurement.P2]
			if !ok1 || !ok2 {
				continue
			}
			result += fmt.Sprintf("l %d %d\n", indexP1, indexP2)

		}
	}

	return []byte(result)
}
