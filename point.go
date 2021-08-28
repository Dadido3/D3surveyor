// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"time"

	"github.com/vugu/vgrouter"
)

type Point struct {
	vgrouter.NavigatorRef `json:"-"`

	site *Site
	key  string

	Name      string
	CreatedAt time.Time

	Position Coordinate
	Optimize bool
}

func (s *Site) NewPoint(name string) *Point {
	key := s.shortIDGen.MustGenerate()

	point := &Point{
		site:      s,
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
	delete(p.site.Points, p.Key())
}

/*func (p *Point) UnmarshalJSON(data []byte) error {
	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Point
	if err := json.Unmarshal(data, tempType(p)); err != nil {
		return err
	}

	// Restore keys and references.

	return nil
}*/

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (p *Point) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	return p.Position.GetTweakablesAndResiduals()
}
