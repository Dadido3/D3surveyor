package main

// Angle describes a angle in radian.
type Angle float64

// TweakableValue returns the values mapped into optimizer space.
func (d *Angle) TweakableValue() float64 {
	return float64(*d)
}

// SetTweakableValue converts and applies the given value from optimizer space.
func (d *Angle) SetTweakableValue(v float64) {
	*d = Angle(v)
}
