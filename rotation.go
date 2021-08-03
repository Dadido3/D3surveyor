package main

// Rotations represents a xyz euler rotation.
type Rotation struct {
	X, Y, Z             Angle
	LockX, LockY, LockZ bool // Lock (Don't optimize) the value.
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (r *Rotation) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables := make([]Tweakable, 0, 3)
	if !r.LockX {
		tweakables = append(tweakables, &r.X)
	}
	if !r.LockY {
		tweakables = append(tweakables, &r.Y)
	}
	if !r.LockZ {
		tweakables = append(tweakables, &r.Z)
	}

	return tweakables, nil
}
