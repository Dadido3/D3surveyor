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
	"time"

	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
)

type TripodMeasurement struct {
	vgrouter.NavigatorRef `json:"-"`

	tripod *Tripod
	key    string

	CreatedAt time.Time

	PointKey         string
	MeasuredDistance Distance // Measured distance between the point and the tripod pivot point.
}

func (t *Tripod) NewMeasurement() *TripodMeasurement {
	tm := new(TripodMeasurement)
	tm.initData()
	tm.initReferences(t, t.site.shortIDGen.MustGenerate())

	// If possible suggest a point mapping.
	if suggestedPoint := t.SuggestPoint(); suggestedPoint != nil {
		tm.PointKey = suggestedPoint.Key()
	}

	return tm
}

// initData initializes the object with default values and other stuff.
func (tm *TripodMeasurement) initData() {
	tm.CreatedAt = time.Now()
}

// initReferences updates references from and to this object and its key.
// This is only used internally to update references for copies or marshalled objects.
// This can't be used on its own to transfer an object from one parent to another.
func (tm *TripodMeasurement) initReferences(newParent *Tripod, newKey string) {
	tm.tripod, tm.key = newParent, newKey
	tm.tripod.Measurements[tm.Key()] = tm
}

func (tm *TripodMeasurement) Key() string {
	return tm.key
}

// DisplayName returns either the name, or if that is empty the key.
func (tm *TripodMeasurement) DisplayName() string {
	return "(" + tm.Key() + ")"
}

func (tm *TripodMeasurement) Delete() {
	delete(tm.tripod.Measurements, tm.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (tm *TripodMeasurement) Copy(newParent *Tripod, newKey string) *TripodMeasurement {
	copy := new(TripodMeasurement)
	copy.initData()
	copy.initReferences(newParent, newKey)
	copy.CreatedAt = tm.CreatedAt
	copy.PointKey = tm.PointKey
	copy.MeasuredDistance = tm.MeasuredDistance

	return copy
}

func (tm *TripodMeasurement) UnmarshalJSON(data []byte) error {
	tm.initData()

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *TripodMeasurement
	if err := json.Unmarshal(data, tempType(tm)); err != nil {
		return err
	}

	// Update parent references and keys.

	return nil
}

func (tm *TripodMeasurement) handleNextSuggestion(event vugu.DOMEvent) {
	if tm.PointKey != "" {
		tm.tripod.ignoredPoints = append(tm.tripod.ignoredPoints, tm.PointKey)
	}

	if suggestedPoint := tm.tripod.SuggestPoint(); suggestedPoint != nil {
		tm.PointKey = suggestedPoint.Key()
	}
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (tm *TripodMeasurement) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	return nil, []Residualer{tm}
}

// ResidualSqr returns the sum of squared residuals. (Each residual is divided by the accuracy of the measurement device).
func (tm *TripodMeasurement) ResidualSqr() float64 {
	tripod := tm.tripod
	site := tripod.site

	if point, ok := site.Points[tm.PointKey]; ok {
		// Determine distance, add offset
		directDistance := tm.MeasuredDistance + tripod.Offset
		pivotDistance := Distance(math.Sqrt(directDistance.Sqr() + tripod.OffsetSide.Sqr()))
		return ((pivotDistance - point.Position.Distance(tripod.Position.Coordinate)) / tripod.Accuracy).Sqr()
	}

	return 0
}
