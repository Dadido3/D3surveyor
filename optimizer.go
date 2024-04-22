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
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/gonum/optimize"
)

// Tweakable is implemented by objects which can be modified in the optimization process.
type Tweakable interface {
	TweakableValue() float64     // TweakableValue returns the values mapped into optimizer space.
	SetTweakableValue(v float64) // SetTweakableValue converts and applies the given value from optimizer space.
}

// TweakableFloat is a optimizable float in the range of -inf to +inf.
type TweakableFloat float64

func (t TweakableFloat) TweakableValue() float64 {
	return float64(t)
}

func (t *TweakableFloat) SetTweakableValue(v float64) {
	*t = TweakableFloat(v)
}

// InputValue implements the valuer interface of the general input component.
func (t TweakableFloat) InputValue() string {
	return fmt.Sprintf("%.13g", t)
}

// SetInputValue implements the valuer interface of the general input component.
func (t *TweakableFloat) SetInputValue(strVal string) {
	strVal = strings.ReplaceAll(strVal, ",", ".")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		log.Printf("strconv.ParseFloat() failed: %v", err)
		return
	}

	*t = TweakableFloat(val)
}

// TweakablePositiveFloat is a optimizable float in the range of 0 to +inf.
type TweakablePositiveFloat float64

func (t TweakablePositiveFloat) TweakableValue() float64 {
	return math.Log(float64(t))
}

func (t *TweakablePositiveFloat) SetTweakableValue(v float64) {
	*t = TweakablePositiveFloat(math.Exp(v))
}

// InputValue implements the valuer interface of the general input component.
func (t TweakablePositiveFloat) InputValue() string {
	return fmt.Sprintf("%.13g", t)
}

// SetInputValue implements the valuer interface of the general input component.
func (t *TweakablePositiveFloat) SetInputValue(strVal string) {
	strVal = strings.ReplaceAll(strVal, ",", ".")

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		log.Printf("strconv.ParseFloat() failed: %v", err)
		return
	}

	*t = TweakablePositiveFloat(val)
}

// Residualer is implemented by objects that can have residuals of measurements or constraints.
type Residualer interface {
	ResidualSqr() float64 // Returns the sum of squared residuals. (Each residual is divided by the accuracy of the measurement device).
}

func Optimize(site *Site, stopFunc func() bool) error {
	tweakables, residuals := site.GetTweakablesAndResiduals()

	if len(tweakables) == 0 {
		return fmt.Errorf("there are no tweakable variables")
	}
	if len(residuals) == 0 {
		return fmt.Errorf("there are no residuals to be determined")
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Function to optimize.
	optimizeFunc := func(x []float64) float64 {
		// Do some silly sleep every now and then to prevent the UI from locking up.
		// TODO: Remove optimizer sleep once WASM threads are fully supported
		select {
		case <-ticker.C:
			time.Sleep(1 * time.Millisecond)
		default:
		}

		site.Lock()
		defer site.Unlock()

		// Set tweakable values.
		for i, tweakable := range tweakables {
			tweakable.SetTweakableValue(x[i])
		}

		// Get sum of squared residuals.
		ssr := 0.0
		for _, residual := range residuals {
			ssr += residual.ResidualSqr()
		}

		return ssr
	}

	// Function to end the optimization prematurely.
	statusFunc := func() (optimize.Status, error) {
		if stopFunc() {
			return optimize.Success, nil
		}

		return optimize.NotTerminated, nil
	}

	p := optimize.Problem{
		Func:   optimizeFunc,
		Status: statusFunc,
	}

	// Get the initial tweakable variables/parameters.
	init := make([]float64, 0, len(tweakables))
	for _, tweakable := range tweakables {
		init = append(init, tweakable.TweakableValue())
		//init = append(init, rand.Float64())
	}

	//res, err := optimize.Minimize(p, init, nil, &optimize.CmaEsChol{InitStepSize: 0.01})
	res, err := optimize.Minimize(p, init, &optimize.Settings{Converger: &optimize.FunctionConverge{Absolute: 1e-10, Iterations: 100000}}, &optimize.NelderMead{})
	if err != nil {
		log.Printf("Optimization failed: %v", err)
	}
	if err = res.Status.Err(); err != nil {
		log.Printf("Optimization status error: %v", err)
	}

	log.Println(res.F, res.X, res.FuncEvaluations, res.MajorIterations)

	// Set tweakable values to the solution.
	for i, tweakable := range tweakables {
		tweakable.SetTweakableValue(res.X[i])
	}

	return nil
}
