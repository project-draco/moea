package nsga

import (
	"testing"

	"project-draco.io/moea"
	"project-draco.io/moea/integer"
)

func TestNewFromString(t *testing.T) {
	ns := &NsgaSelection{
		Variables:   func(i int) []float64 { return []float64{1} },
		LowerBounds: []float64{0},
		UpperBounds: []float64{1},
	}
	rng := MockRNG(1.0)
	p := integer.NewRandomIntegerPopulation(2, 1, []integer.Bound{{0, 1}}, rng)
	c := &moea.Config{Population: p, RandomNumberGenerator: rng}
	ns.initialize(c)
	ns.onGeneration(c, [][]float64{{1.0}, {1.0}})
	t.Log(ns)
}

type MockRNG float64

func (r MockRNG) Flip(probability float64) bool {
	return float64(r) < probability
}

func (r MockRNG) FairFlip() bool {
	return float64(r) < 0.5
}

func (r MockRNG) Float64() float64 {
	return float64(r)
}
