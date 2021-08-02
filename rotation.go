package main

type Rotation struct {
	Yaw, Pitch, Roll             Angle
	LockYaw, LockPitch, LockRoll bool // Lock (Don't optimize) the value.
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (r *Rotation) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables := make([]Tweakable, 0, 3)
	if !r.LockYaw {
		tweakables = append(tweakables, &r.Yaw)
	}
	if !r.LockPitch {
		tweakables = append(tweakables, &r.Pitch)
	}
	if !r.LockRoll {
		tweakables = append(tweakables, &r.Roll)
	}

	return tweakables, nil
}
