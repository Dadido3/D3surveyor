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
	"math"
	"sort"
	"time"

	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
)

type Tripod struct {
	vgrouter.NavigatorRef `json:"-"`

	site *Site
	key  string

	Name      string
	CreatedAt time.Time

	Position                       CoordinateOptimizable // Pivot point of the tripod.
	Accuracy                       Distance              // Accuracy of the measurement.
	Offset, OffsetSide             Distance              // Offset of the rangefinder from the pivot point.
	OffsetLocked, OffsetSideLocked bool                  // Prevent the values from being optimized.

	Measurements  map[string]*TripodMeasurement // List of measurements.
	ignoredPoints []string                      // List of point keys that will not be suggested anymore.
}

func (s *Site) NewTripod(name string) *Tripod {
	t := new(Tripod)
	t.initData()
	t.initReferences(s, s.shortIDGen.MustGenerate())
	t.Name = name

	return t
}

// initData initializes the object with default values and other stuff.
func (t *Tripod) initData() {
	t.CreatedAt = time.Now()
	t.Accuracy = 0.01
	t.OffsetLocked = false
	t.OffsetSideLocked = true
	t.Measurements = map[string]*TripodMeasurement{}
}

// initReferences updates references from and to this object and its key.
// This is only used internally to update references for copies or marshalled objects.
// This can't be used on its own to transfer an object from one parent to another.
func (t *Tripod) initReferences(newParent *Site, newKey string) {
	t.site, t.key = newParent, newKey
	t.site.Tripods[t.Key()] = t
}

func (t *Tripod) handleAdd(event vugu.DOMEvent) {
	measurement := t.NewMeasurement()

	t.Navigate("/tripod/"+t.Key()+"/measurement/"+measurement.Key(), nil)
}

func (t *Tripod) Key() string {
	return t.key
}

// DisplayName returns either the name, or if that is empty the key.
func (t *Tripod) DisplayName() string {
	if t.Name != "" {
		return t.Name
	}

	return "(" + t.Key() + ")"
}

func (t *Tripod) Delete() {
	delete(t.site.Tripods, t.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (t *Tripod) Copy(newParent *Site, newKey string) *Tripod {
	copy := new(Tripod)
	copy.initData()
	copy.initReferences(newParent, newKey)
	copy.Name = t.Name
	copy.CreatedAt = t.CreatedAt
	copy.Position = t.Position
	copy.Accuracy = t.Accuracy
	copy.Offset = t.Offset
	copy.OffsetSide = t.OffsetSide
	copy.OffsetLocked = t.OffsetLocked
	copy.OffsetSideLocked = t.OffsetSideLocked

	// Generate copies of all children.
	for k, v := range t.Measurements {
		v.Copy(copy, k)
	}

	return copy
}

func (t *Tripod) UnmarshalJSON(data []byte) error {
	t.initData()

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Tripod
	if err := json.Unmarshal(data, tempType(t)); err != nil {
		return err
	}

	// Update parent references and keys.
	for k, v := range t.Measurements {
		v.initReferences(t, k)
	}

	return nil
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (t *Tripod) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}

	if !t.OffsetLocked {
		tweakables = append(tweakables, &t.Offset)
	}
	if !t.OffsetSideLocked {
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

// PointUsed returns whether the point is already used in some measurement.
func (t *Tripod) PointUsed(point *Point) bool {
	if point == nil {
		return false
	}

	for _, measurement := range t.Measurements {
		if measurement.PointKey == point.Key() {
			return true
		}
	}

	return false
}

// PointSuggestionAllowed returns whether the point is allowed to be used as suggestion.
func (t *Tripod) PointSuggestionAllowed(point *Point) bool {
	if point == nil {
		return false
	}

	for _, pointKey := range t.ignoredPoints {
		if pointKey == point.Key() {
			return false
		}
	}

	return true
}

// SuggestPoint returns a point that should be measured next.
func (t *Tripod) SuggestPoint() *Point {
	// Go through all measurements, beginning from the newest.
	// Search for a measurement with a valid point, use that point as "previous point".
	for _, measurement := range t.MeasurementsSorted() {
		if previousPoint, ok := t.site.Points[measurement.PointKey]; ok {

			// Find unused point that is closest to the previous point.
			closestPoint, closestDist := (*Point)(nil), Distance(math.Inf(1))
			for _, point := range t.site.Points {
				dist := point.Position.Distance(previousPoint.Position.Coordinate) // TODO: Use squared distance here
				if closestDist > dist && t.PointSuggestionAllowed(point) && !t.PointUsed(point) {
					closestDist, closestPoint = dist, point
				}
			}

			return closestPoint
		}
	}

	return nil
}
