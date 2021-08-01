package main

import (
	"math"
	"time"

	"github.com/vugu/vgrouter"
)

type RangefinderMeasurement struct {
	vgrouter.NavigatorRef

	rangefinder *Rangefinder
	key         string

	CreatedAt time.Time

	P1, P2           string   // Two points the distance is measured between.
	MeasuredDistance Distance // Measured distance in metres.
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

	return math.Pow(float64((p1.Position.Distance(p2.Position)-d.MeasuredDistance)/d.rangefinder.Accuracy), 2) // TODO: Check if this can be optimized
}
