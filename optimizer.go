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
	"log"
	"time"

	"github.com/vugu/vugu"
	"gonum.org/v1/gonum/optimize"
)

// Tweakable is implemented by objects which can be modified in the optimization process.
type Tweakable interface {
	TweakableValue() float64     // TweakableValue returns the values mapped into optimizer space.
	SetTweakableValue(v float64) // SetTweakableValue converts and applies the given value from optimizer space.
}

// Residualer is implemented by objects that can have residuals of measurements or constraints.
type Residualer interface {
	ResidualSqr() float64 // Returns the sum of squared residuals. (Each residual is divided by the accuracy of the measurement device).
}

func Optimize(eventEnv vugu.EventEnv, site *Site) {
	tweakables, residuals := site.GetTweakablesAndResiduals()

	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	f := func(x []float64) float64 {
		// Do some silly UI drawing syncing.
		eventEnv.Lock()
		select {
		case <-ticker.C:
			defer func() {
				eventEnv.UnlockRender()
				time.Sleep(40 * time.Millisecond)
			}()
		default:
			defer eventEnv.UnlockOnly()
		}

		// TODO: Optimize a copy of the site, to prevent locking the UI.

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

	p := optimize.Problem{
		Func: f,
	}

	// Get the initial tweakable variables/parameters.
	init := make([]float64, 0, len(tweakables))
	for _, tweakable := range tweakables {
		init = append(init, tweakable.TweakableValue())
		//init = append(init, rand.Float64())
	}

	/*res, err := optimize.Minimize(p, init, nil, &optimize.CmaEsChol{InitStepSize: 0.01})
	if err != nil {
		log.Printf("Optimization failed: %v", err)
	}
	if err = res.Status.Err(); err != nil {
		log.Printf("Optimization status error: %v", err)
	}*/

	res, err := optimize.Minimize(p, init, &optimize.Settings{Converger: &optimize.FunctionConverge{Absolute: 1e-10, Iterations: 1000}}, &optimize.NelderMead{})
	if err != nil {
		log.Printf("Optimization failed: %v", err)
	}
	if err = res.Status.Err(); err != nil {
		log.Printf("Optimization status error: %v", err)
	}

	log.Println(res.F, res.X, res.FuncEvaluations, res.MajorIterations)

	eventEnv.Lock()
	defer eventEnv.UnlockRender()

	// Set tweakable values to the solution.
	for i, tweakable := range tweakables {
		tweakable.SetTweakableValue(res.X[i])
	}
}
