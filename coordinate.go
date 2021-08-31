// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import "math"

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
