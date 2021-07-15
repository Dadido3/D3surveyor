package main

import (
	"time"

	"github.com/vugu/vgrouter"
)

type Rangefinder struct {
	vgrouter.NavigatorRef
	site *Site
	key  string

	Name      string
	CreatedAt time.Time

	Accuracy float64 // Accuracy of the measurement in metres.

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
