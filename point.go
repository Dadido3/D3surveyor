// Copyright (C) 2021-2025 David Vogel
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
	"time"

	"github.com/vugu/vgrouter"
)

type Point struct {
	vgrouter.NavigatorRef `json:"-"`

	site *Site
	key  string

	Name      string
	CreatedAt time.Time

	Position CoordinateOptimizable
}

func (s *Site) NewPoint(name string) *Point {
	p := new(Point)
	p.initData()
	p.initReferences(s, s.shortIDGen.MustGenerate())
	p.Name = name

	return p
}

// initData initializes the object with default values and other stuff.
func (p *Point) initData() {
	p.CreatedAt = time.Now()
}

// initReferences updates references from and to this object and its key.
// This is only used internally to update references for copies or marshalled objects.
// This can't be used on its own to transfer an object from one parent to another.
func (p *Point) initReferences(newParent *Site, newKey string) {
	p.site, p.key = newParent, newKey
	p.site.Points[p.Key()] = p
}

func (p *Point) Key() string {
	return p.key
}

// DisplayName returns either the name, or if that is empty the key.
func (p *Point) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}

	return "(" + p.Key() + ")"
}

// Delete removes the parent's reference to this object.
func (p *Point) Delete() {
	delete(p.site.Points, p.Key())
}

// Copy returns a copy of the given object.
// Expensive data like images will not be copied, but referenced.
func (p *Point) Copy(newParent *Site, newKey string) *Point {
	copy := new(Point)
	copy.initData()
	copy.initReferences(newParent, newKey)
	copy.Name = p.Name
	copy.CreatedAt = p.CreatedAt
	copy.Position = p.Position

	return copy
}

func (p *Point) UnmarshalJSON(data []byte) error {
	p.initData()

	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Point
	if err := json.Unmarshal(data, tempType(p)); err != nil {
		return err
	}

	// Update parent references and keys.

	return nil
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (p *Point) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	return p.Position.GetTweakablesAndResiduals()
}

// CameraPhotoMappings returns a list of all non suggested mappings related to this point.
func (p *Point) CameraPhotoMappings() []*CameraPhotoMapping {
	mappings := make([]*CameraPhotoMapping, 0)

	for _, camera := range p.site.CamerasSorted() {
		for _, photo := range camera.PhotosSorted() {
			for _, mapping := range photo.MappingsSorted() {
				if !mapping.Suggested && mapping.PointKey == p.key {
					mappings = append(mappings, mapping)
				}
			}
		}
	}

	return mappings
}

// Lines returns a list of all lines that are related to this point.
func (p *Point) Lines() []*Line {
	lines := make([]*Line, 0)

	for _, line := range p.site.LinesSorted() {
		if line.P1 == p.key || line.P2 == p.key {
			lines = append(lines, line)
		}
	}

	return lines
}

// RangefinderMeasurements returns a list of all Rangefinder measurements that are related to this point.
func (p *Point) RangefinderMeasurements() []*RangefinderMeasurement {
	measurements := make([]*RangefinderMeasurement, 0)

	for _, rangefinder := range p.site.RangefindersSorted() {
		for _, measurement := range rangefinder.MeasurementsSorted() {
			if measurement.P1 == p.key || measurement.P2 == p.key {
				measurements = append(measurements, measurement)
			}
		}
	}

	return measurements
}

// TripodMeasurements returns a list of all tripod measurements that are related to this point.
func (p *Point) TripodMeasurements() []*TripodMeasurement {
	measurements := make([]*TripodMeasurement, 0)

	for _, tripod := range p.site.TripodsSorted() {
		for _, measurement := range tripod.MeasurementsSorted() {
			if measurement.PointKey == p.key {
				measurements = append(measurements, measurement)
			}
		}
	}

	return measurements
}
