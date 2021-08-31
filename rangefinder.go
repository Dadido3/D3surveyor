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
	key := s.shortIDGen.MustGenerate()

	r := &Rangefinder{
		site:         s,
		key:          key,
		Name:         name,
		CreatedAt:    time.Now(),
		Accuracy:     0.01,
		Measurements: map[string]*RangefinderMeasurement{},
	}

	s.Rangefinders[key] = r

	return r
}

func (r *Rangefinder) handleAdd() {
	measurement := r.NewMeasurement()

	r.Navigate("/rangefinder/"+r.Key()+"/measurement/"+measurement.Key(), nil)
}

func (r *Rangefinder) Key() string {
	return r.key
}

func (r *Rangefinder) Delete() {
	delete(r.site.Rangefinders, r.Key())
}

func (r *Rangefinder) UnmarshalJSON(data []byte) error {
	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Rangefinder
	if err := json.Unmarshal(data, tempType(r)); err != nil {
		return err
	}

	// Restore keys and references.
	for k, v := range r.Measurements {
		v.key, v.rangefinder = k, r
	}

	return nil
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (r *Rangefinder) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}
	for _, measurement := range r.Measurements {
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
