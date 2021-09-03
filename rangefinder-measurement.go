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
	key := r.site.shortIDGen.MustGenerate()

	rm := &RangefinderMeasurement{
		rangefinder: r,
		key:         key,
		CreatedAt:   time.Now(),
	}

	r.Measurements[key] = rm

	return rm
}

func (d *RangefinderMeasurement) Key() string {
	return d.key
}

func (d *RangefinderMeasurement) Delete() {
	delete(d.rangefinder.Measurements, d.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (d *RangefinderMeasurement) Copy() *RangefinderMeasurement {
	return &RangefinderMeasurement{
		CreatedAt:        d.CreatedAt,
		P1:               d.P1,
		P2:               d.P2,
		MeasuredDistance: d.MeasuredDistance,
	}
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (d *RangefinderMeasurement) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	return nil, []Residualer{d}
}

// ResidualSqr returns the sum of squared residuals. (Each residual is divided by the accuracy of the measurement device).
func (d *RangefinderMeasurement) ResidualSqr() float64 {
	site := d.rangefinder.site

	p1, ok := site.Points[d.P1]
	if !ok {
		return 0
	}
	p2, ok := site.Points[d.P2]
	if !ok {
		return 0
	}

	return sqr(float64((p1.Position.Distance(p2.Position) - d.MeasuredDistance) / d.rangefinder.Accuracy)) // TODO: Check if this can be optimized
}
