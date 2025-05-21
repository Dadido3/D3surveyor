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
	"sort"
	"time"

	"github.com/vugu/vgrouter"
)

type Rangefinder struct {
	vgrouter.NavigatorRef `json:"-"`

	site *Site
	key  string

	Name      string
	CreatedAt time.Time

	Accuracy Distance // Accuracy of the measurement.

	Measurements map[string]*RangefinderMeasurement // List of measurements.
}

func (s *Site) NewRangefinder(name string) *Rangefinder {
	r := new(Rangefinder)
	r.initData()
	r.initReferences(s, s.shortIDGen.MustGenerate())
	r.Name = name

	return r
}

// initData initializes the object with default values and other stuff.
func (r *Rangefinder) initData() {
	r.CreatedAt = time.Now()
	r.Accuracy = 0.01
	r.Measurements = map[string]*RangefinderMeasurement{}
}

// initReferences updates references from and to this object and its key.
// This is only used internally to update references for copies or marshalled objects.
// This can't be used on its own to transfer an object from one parent to another.
func (r *Rangefinder) initReferences(newParent *Site, newKey string) {
	r.site, r.key = newParent, newKey
	r.site.Rangefinders[r.Key()] = r
}

func (r *Rangefinder) handleAdd() {
	measurement := r.NewMeasurement()

	r.Navigate("/rangefinder/"+r.Key()+"/measurement/"+measurement.Key(), nil)
}

func (r *Rangefinder) Key() string {
	return r.key
}

// DisplayName returns either the name, or if that is empty the key.
func (r *Rangefinder) DisplayName() string {
	if r.Name != "" {
		return r.Name
	}

	return "(" + r.Key() + ")"
}

func (r *Rangefinder) Delete() {
	delete(r.site.Rangefinders, r.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (r *Rangefinder) Copy(newParent *Site, newKey string) *Rangefinder {
	copy := new(Rangefinder)
	copy.initData()
	copy.initReferences(newParent, newKey)
	copy.Name = r.Name
	copy.CreatedAt = r.CreatedAt
	copy.Accuracy = r.Accuracy

	// Generate copies of all children.
	for k, v := range r.Measurements {
		v.Copy(copy, k)
	}

	return copy
}

func (r *Rangefinder) UnmarshalJSON(data []byte) error {
	r.initData()

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Rangefinder
	if err := json.Unmarshal(data, tempType(r)); err != nil {
		return err
	}

	// Update parent references and keys.
	for k, v := range r.Measurements {
		v.initReferences(r, k)
	}

	return nil
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (r *Rangefinder) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}
	for _, measurement := range r.MeasurementsSorted() {
		newTweakables, newResiduals := measurement.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}
	return tweakables, residuals
}

// MeasurementsSorted returns the measurements of the rangefinder as a list sorted by date.
// TODO: Replace with generics once they are available. It's one of the few cases where they are really needed
func (s *Rangefinder) MeasurementsSorted() []*RangefinderMeasurement {
	measurements := make([]*RangefinderMeasurement, 0, len(s.Measurements))

	for _, measurement := range s.Measurements {
		measurements = append(measurements, measurement)
	}

	sort.Slice(measurements, func(i, j int) bool {
		return measurements[i].CreatedAt.After(measurements[j].CreatedAt)
	})

	return measurements
}
