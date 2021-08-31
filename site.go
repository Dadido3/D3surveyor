// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"sort"

	"github.com/teris-io/shortid"
	"github.com/vugu/vgrouter"
)

// Site is the root container for all measurements, points and constraints of a place/site.
type Site struct {
	vgrouter.NavigatorRef `json:"-"`

	shortIDGen *shortid.Shortid

	Name string

	// Geometry data and measurements.
	Points       map[string]*Point
	Cameras      map[string]*Camera
	Rangefinders map[string]*Rangefinder
	Tripods      map[string]*Tripod
}

func NewSite(name string) (*Site, error) {
	shortIDGen, err := shortid.New(0, shortid.DefaultABC, 1234)
	if err != nil {
		return nil, err
	}

	site := &Site{
		shortIDGen:   shortIDGen,
		Name:         name,
		Points:       map[string]*Point{},
		Cameras:      map[string]*Camera{},
		Rangefinders: map[string]*Rangefinder{},
		Tripods:      map[string]*Tripod{},
	}

	return site, nil
}

func MustNewSite(name string) *Site {
	site, err := NewSite(name)
	if err != nil {
		panic(err)
	}

	return site
}

func NewSiteFromJSON(data []byte) (*Site, error) {
	site := &Site{}
	if err := json.Unmarshal(data, site); err != nil {
		return nil, err
	}

	return site, nil
}

func (s *Site) UnmarshalJSON(data []byte) error {
	newSite, err := NewSite("")
	if err != nil {
		return err
	}

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Site
	if err := json.Unmarshal(data, tempType(newSite)); err != nil {
		return err
	}

	// Restore keys and references.
	for k, v := range newSite.Points {
		v.key, v.site = k, s
	}
	for k, v := range newSite.Cameras {
		v.key, v.site = k, s
	}
	for k, v := range newSite.Rangefinders {
		v.key, v.site = k, s
	}
	for k, v := range newSite.Tripods {
		v.key, v.site = k, s
	}

	// Overwrite data of existing site.
	*s = *newSite
	return nil
}

// Global site data structure that contains all data about a specific site/place.
var globalSite *Site = MustNewSite("New")

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (s *Site) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}

	for _, point := range s.Points {
		newTweakables, newResiduals := point.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	for _, camera := range s.Cameras {
		newTweakables, newResiduals := camera.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	for _, rangefinder := range s.Rangefinders {
		newTweakables, newResiduals := rangefinder.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	for _, tripod := range s.Tripods {
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
