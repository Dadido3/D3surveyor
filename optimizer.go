package main

import (
	"log"

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

func Optimize(site *Site) {
	tweakables, residuals := site.GetTweakablesAndResiduals()

	f := func(x []float64) float64 {
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
	}

	res, err := optimize.Minimize(p, init, nil, &optimize.CmaEsChol{})
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
}
