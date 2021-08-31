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
