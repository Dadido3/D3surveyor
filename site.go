package main

import (
	"github.com/teris-io/shortid"
	"github.com/vugu/vgrouter"
)

type Site struct {
	vgrouter.NavigatorRef

	shortIDGen *shortid.Shortid

	Name string

	Points map[string]*Point

	// Capture and measurement devices.
	//Photos map[string]*Photo
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
