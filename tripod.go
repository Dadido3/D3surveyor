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

type Tripod struct {
	vgrouter.NavigatorRef `json:"-"`

	site *Site
	key  string

	Name      string
	CreatedAt time.Time

	Position                   Coordinate // Pivot point of the tripod.
	Accuracy                   Distance   // Accuracy of the measurement.
	Offset, OffsetSide         Distance   // Offset of the rangefinder from the pivot point.
	OffsetLock, OffsetSideLock bool       // Prevent the values from being optimized.

	Measurements map[string]*TripodMeasurement // List of measurements.
}

func (s *Site) NewTripod(name string) *Tripod {
	key := s.shortIDGen.MustGenerate()

	r := &Tripod{
		site:         s,
		key:          key,
		Name:         name,
		CreatedAt:    time.Now(),
		Accuracy:     0.01,
		Measurements: map[string]*TripodMeasurement{},
	}

	s.Tripods[key] = r

	return r
}

func (t *Tripod) handleAdd() {
	measurement := t.NewMeasurement()

	t.Navigate("/tripod/"+t.Key()+"/measurement/"+measurement.Key(), nil)
}

func (t *Tripod) Key() string {
	return t.key
}

func (t *Tripod) Delete() {
	delete(t.site.Tripods, t.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (t *Tripod) Copy() *Tripod {
	copy := &Tripod{
		Name:           t.Name,
		CreatedAt:      t.CreatedAt,
		Position:       t.Position,
		Accuracy:       t.Accuracy,
		Offset:         t.Offset,
		OffsetSide:     t.OffsetSide,
		OffsetLock:     t.OffsetLock,
		OffsetSideLock: t.OffsetSideLock,
		Measurements:   map[string]*TripodMeasurement{},
	}

	// Generate copies of all children.
	for k, v := range t.Measurements {
		copy.Measurements[k] = v.Copy()
	}

	// Restore keys and references.
	copy.RestoreChildrenRefs()

	return copy
}

// RestoreChildrenRefs updates the key of the children and any reference to this object.
func (t *Tripod) RestoreChildrenRefs() {
	for k, v := range t.Measurements {
		v.key, v.tripod = k, t
	}
}

func (t *Tripod) UnmarshalJSON(data []byte) error {
	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Tripod
	if err := json.Unmarshal(data, tempType(t)); err != nil {
		return err
	}

	// Restore keys and references.
	t.RestoreChildrenRefs()

	return nil
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (t *Tripod) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}

	if !t.OffsetLock {
		tweakables = append(tweakables, &t.Offset)
	}
	if !t.OffsetSideLock {
		tweakables = append(tweakables, &t.OffsetSide)
	}

	for _, measurement := range t.MeasurementsSorted() {
		newTweakables, newResiduals := measurement.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	newTweakables, newResiduals := t.Position.GetTweakablesAndResiduals()
	tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)

	return tweakables, residuals
}

// MeasurementsSorted returns the measurements of the tripod as a list sorted by date.
// TODO: Replace with generics once they are available. It's one of the few cases where they are really needed
func (t *Tripod) MeasurementsSorted() []*TripodMeasurement {
	measurements := make([]*TripodMeasurement, 0, len(t.Measurements))

	for _, measurement := range t.Measurements {
		measurements = append(measurements, measurement)
	}

	sort.Slice(measurements, func(i, j int) bool {
		return measurements[i].CreatedAt.After(measurements[j].CreatedAt)
	})

	return measurements
}
