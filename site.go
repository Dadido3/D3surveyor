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
	"sort"
	"sync"

	"github.com/teris-io/shortid"
)

// Site is the root container for all measurements, points and constraints of a place/site.
type Site struct {
	sync.RWMutex `json:"-"`

	shortIDGen *shortid.Shortid

	optimizerState OptimizerState

	Name string

	// Geometry data and measurements.
	Points       map[string]*Point
	Lines        map[string]*Line
	Cameras      map[string]*Camera
	Rangefinders map[string]*Rangefinder
	Tripods      map[string]*Tripod
}

func NewSite(name string) *Site {
	s := new(Site)
	s.initData()
	s.Name = name

	return s
}

func NewSiteFromJSON(data []byte) (*Site, error) {
	site := new(Site)

	// The site will be initialized by UnmarshalJSON.
	if err := json.Unmarshal(data, site); err != nil {
		return nil, err
	}

	return site, nil
}

// Initialize the object with default values and other stuff.
func (s *Site) initData() {
	s.shortIDGen = shortid.MustNew(0, shortid.DefaultABC, 1234)
	s.optimizerState.site = s
	s.Points = map[string]*Point{}
	s.Lines = map[string]*Line{}
	s.Cameras = map[string]*Camera{}
	s.Rangefinders = map[string]*Rangefinder{}
	s.Tripods = map[string]*Tripod{}
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (s *Site) Copy() *Site {
	copy := new(Site)
	copy.initData()
	//copy.updateReferences(newParent, newKey)
	copy.Name = s.Name

	// Generate copies of all children. Also update their parent reference and key.
	for k, v := range s.Points {
		v.Copy(copy, k)
	}
	for k, v := range s.Lines {
		v.Copy(copy, k)
	}
	for k, v := range s.Cameras {
		v.Copy(copy, k)
	}
	for k, v := range s.Rangefinders {
		v.Copy(copy, k)
	}
	for k, v := range s.Tripods {
		v.Copy(copy, k)
	}

	return copy
}

func (s *Site) UnmarshalJSON(data []byte) error {
	s.initData()

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Site
	if err := json.Unmarshal(data, tempType(s)); err != nil {
		return err
	}

	// Update parent references and keys.
	for k, v := range s.Points {
		v.initReferences(s, k)
	}
	for k, v := range s.Lines {
		v.initReferences(s, k)
	}
	for k, v := range s.Cameras {
		v.initReferences(s, k)
	}
	for k, v := range s.Rangefinders {
		v.initReferences(s, k)
	}
	for k, v := range s.Tripods {
		v.initReferences(s, k)
	}

	return nil
}

// Global site data structure that contains all data about a specific site/place.
var globalSite *Site = NewSite("New")

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (s *Site) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}

	for _, point := range s.PointsSorted() {
		newTweakables, newResiduals := point.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	for _, line := range s.LinesSorted() {
		newTweakables, newResiduals := line.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	for _, camera := range s.CamerasSorted() {
		newTweakables, newResiduals := camera.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	for _, rangefinder := range s.RangefindersSorted() {
		newTweakables, newResiduals := rangefinder.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	for _, tripod := range s.TripodsSorted() {
		newTweakables, newResiduals := tripod.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	return tweakables, residuals
}

// PointsSorted returns the points of the site as a list sorted by date.
// TODO: Replace with generics once they are available. It's one of the few cases where they are really needed
func (s *Site) PointsSorted() []*Point {
	points := make([]*Point, 0, len(s.Points))

	for _, point := range s.Points {
		points = append(points, point)
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].CreatedAt.After(points[j].CreatedAt)
	})

	return points
}

// LinesSorted returns the lines of the site as a list sorted by date.
// TODO: Replace with generics once they are available. It's one of the few cases where they are really needed
func (s *Site) LinesSorted() []*Line {
	lines := make([]*Line, 0, len(s.Lines))

	for _, line := range s.Lines {
		lines = append(lines, line)
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].CreatedAt.After(lines[j].CreatedAt)
	})

	return lines
}

// RangefindersSorted returns the rangefinders of the site as a list sorted by date.
// TODO: Replace with generics once they are available. It's one of the few cases where they are really needed
func (s *Site) RangefindersSorted() []*Rangefinder {
	rangefinders := make([]*Rangefinder, 0, len(s.Rangefinders))

	for _, rangefinder := range s.Rangefinders {
		rangefinders = append(rangefinders, rangefinder)
	}

	sort.Slice(rangefinders, func(i, j int) bool {
		return rangefinders[i].CreatedAt.After(rangefinders[j].CreatedAt)
	})

	return rangefinders
}

// CamerasSorted returns the cameras of the site as a list sorted by date.
// TODO: Replace with generics once they are available. It's one of the few cases where they are really needed
func (s *Site) CamerasSorted() []*Camera {
	cameras := make([]*Camera, 0, len(s.Cameras))

	for _, camera := range s.Cameras {
		cameras = append(cameras, camera)
	}

	sort.Slice(cameras, func(i, j int) bool {
		return cameras[i].CreatedAt.After(cameras[j].CreatedAt)
	})

	return cameras
}

// TripodsSorted returns the tripods of the site as a list sorted by date.
// TODO: Replace with generics once they are available. It's one of the few cases where they are really needed
func (s *Site) TripodsSorted() []*Tripod {
	tripods := make([]*Tripod, 0, len(s.Tripods))

	for _, tripod := range s.Tripods {
		tripods = append(tripods, tripod)
	}

	sort.Slice(tripods, func(i, j int) bool {
		return tripods[i].CreatedAt.After(tripods[j].CreatedAt)
	})

	return tripods
}
