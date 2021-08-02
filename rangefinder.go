package main

import (
	"encoding/json"
	"time"

	"github.com/vugu/vgrouter"
)

type Rangefinder struct {
	vgrouter.NavigatorRef `json:"-"`

	site *Site
	key  string

	Name      string
	CreatedAt time.Time

	Accuracy Distance // Accuracy of the measurement in metres.

	Measurements map[string]*RangefinderMeasurement // List of measurements.
}

func (s *Site) NewRangefinder(name string) *Rangefinder {
	key := s.shortIDGen.MustGenerate()

	r := &Rangefinder{
		site:         s,
		key:          key,
		Name:         name,
		CreatedAt:    time.Now(),
		Accuracy:     0.01,
		Measurements: map[string]*RangefinderMeasurement{},
	}

	s.Rangefinders[key] = r

	return r
}

func (r *Rangefinder) handleAdd() {
	measurement := r.NewMeasurement()

	r.Navigate("/rangefinder/"+r.Key()+"/measurement/"+measurement.Key(), nil)
}

func (r *Rangefinder) Key() string {
	return r.key
}

func (r *Rangefinder) Delete() {
	delete(r.site.Rangefinders, r.Key())
}

func (r *Rangefinder) UnmarshalJSON(data []byte) error {
	// Unmarshal structure normally. Cast it into a different type to prevent recursion with json.Unmarshal.
	type tempType *Rangefinder
	if err := json.Unmarshal(data, tempType(r)); err != nil {
		return err
	}

	// Restore keys and references.
	for k, v := range r.Measurements {
		v.key, v.rangefinder = k, r
	}

	return nil
}

// GetTweakablesAndResiduals returns a list of tweakable variables and residuals.
func (r *Rangefinder) GetTweakablesAndResiduals() ([]Tweakable, []Residualer) {
	tweakables, residuals := []Tweakable{}, []Residualer{}
	for _, measurement := range r.Measurements {
		newTweakables, newResiduals := measurement.GetTweakablesAndResiduals()
		tweakables, residuals = append(tweakables, newTweakables...), append(residuals, newResiduals...)
	}
	return tweakables, residuals
}
