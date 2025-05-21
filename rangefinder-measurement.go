// Copyright (C) 2021-2025 David Vogel
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

import (
	"encoding/json"
	"time"

	"github.com/vugu/vgrouter"
)

type RangefinderMeasurement struct {
	vgrouter.NavigatorRef `json:"-"`

	rangefinder *Rangefinder
	key         string

	CreatedAt time.Time

	P1, P2           string   // Two points the distance is measured between.
	MeasuredDistance Distance // Measured distance.
}

func (r *Rangefinder) NewMeasurement() *RangefinderMeasurement {
	rm := new(RangefinderMeasurement)
	rm.initData()
	rm.initReferences(r, r.site.shortIDGen.MustGenerate())

	return rm
}

// initData initializes the object with default values and other stuff.
func (rm *RangefinderMeasurement) initData() {
	rm.CreatedAt = time.Now()
}

// initReferences updates references from and to this object and its key.
// This is only used internally to update references for copies or marshalled objects.
// This can't be used on its own to transfer an object from one parent to another.
func (rm *RangefinderMeasurement) initReferences(newParent *Rangefinder, newKey string) {
	rm.rangefinder, rm.key = newParent, newKey
	rm.rangefinder.Measurements[rm.Key()] = rm
}

func (rm *RangefinderMeasurement) Key() string {
	return rm.key
}

// DisplayName returns either the name, or if that is empty the key.
func (rm *RangefinderMeasurement) DisplayName() string {
	return "(" + rm.Key() + ")"
}

func (rm *RangefinderMeasurement) Delete() {
	delete(rm.rangefinder.Measurements, rm.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (rm *RangefinderMeasurement) Copy(newParent *Rangefinder, newKey string) *RangefinderMeasurement {
	copy := new(RangefinderMeasurement)
	copy.initData()
	copy.initReferences(newParent, newKey)
	copy.CreatedAt = rm.CreatedAt
	copy.P1 = rm.P1
	copy.P2 = rm.P2
	copy.MeasuredDistance = rm.MeasuredDistance

	return copy
}

func (rm *RangefinderMeasurement) UnmarshalJSON(data []byte) error {
	rm.initData()

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *RangefinderMeasurement
	if err := json.Unmarshal(data, tempType(rm)); err != nil {
		return err
	}

	// Update parent references and keys.

	return nil
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (rm *RangefinderMeasurement) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	return nil, []Residualer{rm}
}

// ResidualSqr returns the sum of squared residuals. (Each residual is divided by the accuracy of the measurement device).
func (rm *RangefinderMeasurement) ResidualSqr() float64 {
	site := rm.rangefinder.site

	p1, ok := site.Points[rm.P1]
	if !ok {
		return 0
	}
	p2, ok := site.Points[rm.P2]
	if !ok {
		return 0
	}

	return ((p1.Position.Distance(p2.Position.Coordinate) - rm.MeasuredDistance) / rm.rangefinder.Accuracy).Sqr() // TODO: Check if this can be optimized
}
