package main

import (
	"github.com/teris-io/shortid"
	"github.com/vugu/vgrouter"
)

// Site is the root container for all measurements, points and constraints of a place/site.
type Site struct {
	vgrouter.NavigatorRef

	shortIDGen *shortid.Shortid

	Name string

	Points map[string]*Point

	// Capture and measurement devices.
	Cameras      map[string]*Camera
	Rangefinders map[string]*Rangefinder
	//TripodMeasurements map[string]*TripodMeasurement
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

// Global site data structure that contains all data about a specific site/place.
var globalSite *Site = MustNewSite("test")

func init() {
	globalSite.NewPoint("1")
	globalSite.NewPoint("2")
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (s *Site) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}

	for _, point := range s.Points {
		newTweakables, newResiduals := point.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	for _, rangefinder := range s.Rangefinders {
		newTweakables, newResiduals := rangefinder.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	for _, camera := range s.Cameras {
		newTweakables, newResiduals := camera.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}

	return tweakables, residuals
}
