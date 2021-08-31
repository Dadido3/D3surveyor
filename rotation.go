// Copyright (C) 2021 David Vogel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
