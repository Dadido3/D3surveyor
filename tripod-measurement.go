// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"math"
	"time"

	"github.com/vugu/vgrouter"
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
	key := t.site.shortIDGen.MustGenerate()

	rm := &TripodMeasurement{
		tripod:    t,
		key:       key,
		CreatedAt: time.Now(),
	}

	t.Measurements[key] = rm

	return rm
}

func (tm *TripodMeasurement) Key() string {
	return tm.key
}

func (tm *TripodMeasurement) Delete() {
	delete(tm.tripod.Measurements, tm.Key())
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
		pivotDistance := Distance(math.Sqrt(sqr(float64(directDistance)) + sqr(float64(tripod.OffsetSide))))
		return sqr(float64((pivotDistance - point.Position.Distance(tripod.Position)) / tripod.Accuracy))
	}

	return 0
}
