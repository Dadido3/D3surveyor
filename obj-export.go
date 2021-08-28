// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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
