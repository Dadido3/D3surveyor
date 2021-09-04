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
	"math"
	"time"

	"github.com/vugu/vgrouter"
)

type Line struct {
	vgrouter.NavigatorRef `json:"-"`

	site *Site
	key  string

	CreatedAt time.Time

	P1, P2 string

	// Parallelism to a given vector.
	ParallelEnabled  bool
	ParallelVector   Coordinate
	ParallelAccuracy Angle
}

func (s *Site) NewLine() *Line {
	key := s.shortIDGen.MustGenerate()

	line := &Line{
		site:             s,
		key:              key,
		CreatedAt:        time.Now(),
		ParallelVector:   Coordinate{X: 0, Y: 0, Z: 1},
		ParallelAccuracy: Angle(1 * math.Pi / 180),
	}

	s.Lines[key] = line

	return line
}

func (l *Line) Key() string {
	return l.key
}

func (l *Line) Delete() {
	delete(l.site.Lines, l.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (l *Line) Copy() *Line {
	return &Line{
		CreatedAt:        l.CreatedAt,
		P1:               l.P1,
		P2:               l.P2,
		ParallelEnabled:  l.ParallelEnabled,
		ParallelVector:   l.ParallelVector,
		ParallelAccuracy: l.ParallelAccuracy,
	}
}

/*func (l *Line) UnmarshalJSON(data []byte) error {
	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Line
	if err := json.Unmarshal(data, tempType(l)); err != nil {
		return err
	}

	// Restore keys and references.
	l.RestoreChildrenRefs()

	return nil
}*/

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (l *Line) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	return nil, []Residualer{l}
}

// ResidualSqr returns the sum of squared residuals. (Each residual is divided by the accuracy of the measurement device).
func (l *Line) ResidualSqr() float64 {
	site := l.site

	p1, ok := site.Points[l.P1]
	if !ok {
		return 0
	}
	p2, ok := site.Points[l.P2]
	if !ok {
		return 0
	}

	ssr := 0.0

	if l.ParallelEnabled {
		v1, v2 := l.ParallelVector.Vec3(), p2.Position.Vec3().Sub(p1.Position.Vec3())

		sr := sqr(math.Acos(v1.Dot(v2)/v1.Len()/v2.Len()) / float64(l.ParallelAccuracy))
		if math.IsNaN(sr) {
			sr = 1000000
		}
		ssr += math.Min(sr, 1000000)
	}

	return ssr
}
