package nsgaii

import (
	"testing"

	"project-draco.io/moea"
	"project-draco.io/moea/integer"
)

var n *NsgaIISelection
var c *moea.Config
var rng moea.RNG

func init() {
	n = &NsgaIISelection{}
	rng = moea.NewXorshift()
	p := integer.NewRandomIntegerPopulation(4, 1, []integer.Bound{{0, 10}}, rng)
	c = &moea.Config{Population: p, NumberOfObjectives: 1, RandomNumberGenerator: rng}
	n.Initialize(c)
}

func TestAssignCrowdingDistance(t *testing.T) {
	for _, f := range []struct {
		in  [][]float64
		out []float64
	}{
		{[][]float64{{0.0}, {1.0}, {2.0}, {3.0}}, []float64{2.0 / 3.0, 2.0 / 3.0}},
		{[][]float64{{0.0}, {2.5}, {2.0}, {3.0}}, []float64{1.0 / 3.0, 2.5 / 3.0}},
	} {
		n.assignCrowdingDistance(f.in, []int{0, 1, 2, 3})
		for i := 1; i < 3; i++ {
			if n.crowdingDistance[i] != f.out[i-1] {
				t.Error("Expected ", f.out[i-1], " but was ", n.crowdingDistance[i])
			}
		}
	}
}

func TestCheckDominance(t *testing.T) {
	for _, f := range []struct {
		in  [][]float64
		out int
	}{
		{[][]float64{{0.0}, {1.0}}, 1},
		{[][]float64{{1.0}, {0.0}}, -1},
		{[][]float64{{0.0}, {0.0}}, 0},
		{[][]float64{{0.0, 0.0}, {0.0, 1.0}}, 1},
		{[][]float64{{1.0, 0.0}, {0.0, 0.0}}, -1},
		{[][]float64{{1.0, 0.0}, {0.0, 1.0}}, 0},
	} {
		d := n.checkDominance(f.in, 0, 1)
		if d != f.out {
			t.Error("Expected ", f.out, " but was ", d)
		}
	}
}

func TestCrowdingFill(t *testing.T) {
	newPopulation := integer.NewRandomIntegerPopulation(4, 1, []integer.Bound{{0, 10}}, rng)
	for _, f := range []struct {
		in  [][]float64
		out []int
	}{
		{[][]float64{{0.0}, {1.0}, {2.5}, {3.0}}, []int{2, 1}},
		{[][]float64{{0.0}, {0.25}, {2.5}, {3.0}}, []int{1, 2}},
	} {
		n.crowdingFill(f.in, c.Population, newPopulation, []int{0, 1, 2, 3}, 1)
		for i := 1; i < 3; i++ {
			if c.Population.Individual(i).Value(0) != newPopulation.Individual(f.out[i-1]).Value(0) {
				t.Error("Expected", c.Population.Individual(i).Value(0),
					"but was", newPopulation.Individual(f.out[i-1]).Value(0))
			}
		}
	}
}

func TestFillNondominatedSort(t *testing.T) {
	mixedPopulation := integer.NewRandomIntegerPopulation(8, 1, []integer.Bound{{0, 10}}, rng)
	newPopulation := integer.NewRandomIntegerPopulation(4, 1, []integer.Bound{{0, 10}}, rng)
	for _, f := range []struct {
		in  [][]float64
		out []int
	}{
		{[][]float64{{0.0}, {1.0}, {2.0}, {3.0}}, []int{0, 1, 2, 3}},
		// {[][]float64{{3.0}, {2.0}, {1.0}, {0.0}}, []int{3, 2, 1, 0}},
	} {
		n.fillNondominatedSort(c, f.in, mixedPopulation, newPopulation)
		for i := 0; i < 4; i++ {
			if c.Population.Individual(i).Value(0) != newPopulation.Individual(f.out[i-1]).Value(0) {
				t.Error("Expected", c.Population.Individual(i).Value(0),
					"but was", newPopulation.Individual(f.out[i-1]).Value(0))
			}
		}
	}
}
