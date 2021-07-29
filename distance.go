package main

// Distance describes a distance in meters, or an absolute position measured by its distance from the origin.
type Distance float64

// TweakableValue returns the values mapped into optimizer space.
func (d *Distance) TweakableValue() float64 {
	return float64(*d)
}

// TweakableValue converts and sets the given value from optimizer space.
func (d *Distance) SetTweakableValue(v float64) {
	*d = Distance(v)
}
