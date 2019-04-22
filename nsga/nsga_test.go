package nsga

import (
	"testing"

	"github.com/project-draco/moea"
	"github.com/project-draco/moea/integer"
)

func TestNewFromString(t *testing.T) {
	ns := &NsgaSelection{
		ValuesAsFloat: func(i moea.Individual) []float64 { return []float64{1.0} },
		LowerBounds:   []float64{0},
		UpperBounds:   []float64{1},
	}
	rng := MockRNG(0.5)
	p := integer.NewRandomIntegerPopulation(2, 1, []integer.Bound{{0, 1}}, rng)
	c := &moea.Config{Population: p, RandomNumberGenerator: rng}
	ns.Initialize(c)
	ns.OnGeneration(c, p, [][]float64{{0.5}, {1.0}})
	if ns.Selection(c, nil) != 1 {
		t.Error("First selection must be 1")
	}
	if ns.Selection(c, nil) != 0 {
		t.Error("Second selection must be 0")
	}
	ns.OnGeneration(c, p, [][]float64{{1.0}, {0.5}})
	if ns.Selection(c, nil) != 0 {
		t.Error("First selection must be 1")
	}
	if ns.Selection(c, nil) != 1 {
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
