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
	"encoding/json"
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

	// Fit line along a given direction vector.
	DirectionEnabled  bool
	DirectionVector   Coordinate
	DirectionAccuracy Angle
}

func (s *Site) NewLine() *Line {
	l := new(Line)
	l.initData()
	l.initReferences(s, s.shortIDGen.MustGenerate())

	return l
}

// initData initializes the object with default values and other stuff.
func (l *Line) initData() {
	l.CreatedAt = time.Now()
	l.DirectionVector = Coordinate{0, 0, 1}
	l.DirectionAccuracy = Angle(1 * math.Pi / 180)
}

// initReferences updates references from and to this object and its key.
// This is only used internally to update references for copies or marshalled objects.
// This can't be used on its own to transfer an object from one parent to another.
func (l *Line) initReferences(newParent *Site, newKey string) {
	l.site, l.key = newParent, newKey
	l.site.Lines[l.Key()] = l
}

func (l *Line) Key() string {
	return l.key
}

func (l *Line) Delete() {
	delete(l.site.Lines, l.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (l *Line) Copy(newParent *Site, newKey string) *Line {
	copy := new(Line)
	copy.initData()
	copy.initReferences(newParent, newKey)
	copy.CreatedAt = l.CreatedAt
	copy.P1 = l.P1
	copy.P2 = l.P2
	copy.DirectionEnabled = l.DirectionEnabled
	copy.DirectionVector = l.DirectionVector
	copy.DirectionAccuracy = l.DirectionAccuracy

	return copy
}

func (l *Line) UnmarshalJSON(data []byte) error {
	l.initData()

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Line
	if err := json.Unmarshal(data, tempType(l)); err != nil {
		return err
	}

	// Update parent references and keys.

	return nil
}

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

	if l.DirectionEnabled {
		v1, v2 := l.DirectionVector.Vec3(), p2.Position.Vec3().Sub(p1.Position.Vec3())

		r := math.Acos(v1.Dot(v2)/v1.Len()/v2.Len()) / float64(l.DirectionAccuracy)
		sr := r * r
		if math.IsNaN(sr) {
			sr = 1000000
		}
		ssr += math.Min(sr, 1000000)
	}

	return ssr
}
