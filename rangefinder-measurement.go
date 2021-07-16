package main

import (
	"time"

	"github.com/vugu/vgrouter"
)

type RangefinderMeasurement struct {
	vgrouter.NavigatorRef
	rangefinder *Rangefinder
	key         string

	CreatedAt time.Time

	P1, P2           string  // Two points the distance is measured between.
	MeasuredDistance float64 // Measured distance in metres.
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
