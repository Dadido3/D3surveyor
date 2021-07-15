package main

import (
	"time"
)

type Point struct {
	Site *Site
	key  string

	Name      string
	CreatedAt time.Time

	Position Coordinate
	Optimize bool
}

func (s *Site) NewPoint(name string) *Point {
	key := s.shortIDGen.MustGenerate()

	point := &Point{
		Site:      s,
		key:       key,
		Name:      name,
		CreatedAt: time.Now(),
	}

	s.Points[key] = point

	return point
}

func (p *Point) Key() string {
	return p.key
}

func (p *Point) Delete() {
	delete(p.Site.Points, p.Key())
}
