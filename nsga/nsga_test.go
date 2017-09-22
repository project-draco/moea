package nsga

import (
	"testing"

	"project-draco.io/moea"
	"project-draco.io/moea/integer"
)

func TestNewFromString(t *testing.T) {
	ns := &NsgaSelection{
		Variables:   func(i int) []float64 { return []float64{float64(i)} },
		LowerBounds: []float64{0},
		UpperBounds: []float64{1},
	}
	rng := MockRNG(0.5)
	p := integer.NewRandomIntegerPopulation(2, 1, []integer.Bound{{0, 1}}, rng)
	c := &moea.Config{Population: p, RandomNumberGenerator: rng, NumberOfObjectives: 1}
	ns.initialize(c)
	ns.onGeneration(c, [][]float64{{0.5}, {1.0}})
	if ns.selection(c, nil) != 1 {
		t.Error("First selection must be 1")
	}
	if ns.selection(c, nil) != 0 {
		t.Error("Second selection must be 0")
	}
	ns.onGeneration(c, [][]float64{{1.0}, {0.5}})
	if ns.selection(c, nil) != 0 {
		t.Error("First selection must be 1")
	}
	if ns.selection(c, nil) != 1 {
		t.Error("Second selection must be 0")
	}
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
