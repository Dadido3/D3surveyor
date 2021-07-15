package main

import (
	"time"
)

type DistanceMeasurement struct {
	Site *Site
	key  string

	Name      string
	CreatedAt time.Time

	P1, P2                     string  // Two points the distance is measured between.
	MeasuredDistance, Accuracy float64 // In metres.
}

func (s *Site) NewDistanceMeasurement(name string) *DistanceMeasurement {
	key := s.shortIDGen.MustGenerate()

	md := &DistanceMeasurement{
		Site:      s,
		key:       key,
		Name:      name,
		CreatedAt: time.Now(),
	}

	s.DistanceMeasurements[key] = md

	return md
}

func (d *DistanceMeasurement) Key() string {
	return d.key
}

func (d *DistanceMeasurement) Delete() {
	delete(d.Site.DistanceMeasurements, d.Key())
}
