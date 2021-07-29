package main

// Tweakable is implemented by objects which can be modified in the optimization process.
type Tweakable interface {
	TweakableValue() float64      // TweakableValue returns the values mapped into optimizer space.
	SetVTweakableValue(v float64) // TweakableValue converts and sets the given value from optimizer space.
}

// Residualer is implemented by objects that can have residuals of measurements or constraints.
type Residualer interface {
	ResidualSqr() float64
}
