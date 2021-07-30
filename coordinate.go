package main

import "math"

type Coordinate struct {
	X, Y, Z Distance // In metres.
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (c *Coordinate) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	return []Tweakable{&c.X, &c.Y, &c.Z}, nil
}

// Distance returns the distance between itself and the second coordinate.
func (c *Coordinate) Distance(c2 Coordinate) Distance {
	sqr := (c.X-c2.X)*(c.X-c2.X) + (c.Y-c2.Y)*(c.Y-c2.Y) + (c.Z-c2.Z)*(c.Z-c2.Z)
	return Distance(math.Sqrt(float64(sqr)))
}
