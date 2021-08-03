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
