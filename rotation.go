package main

type Rotation struct {
	Yaw, Pitch, Roll Angle
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (c *Rotation) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	return []Tweakable{&c.Yaw, &c.Pitch, &c.Roll}, nil
}
