package nsgaii

import (
	"math"
	"reflect"
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
	c.NumberOfObjectives = 2
	n.Initialize(c)
	for _, f := range []struct {
		in  [][]float64
		out []float64
	}{
		{[][]float64{{0.0}, {1.0}, {2.0}, {3.0}}, []float64{2.0 / 3.0, 2.0 / 3.0}},
		{[][]float64{{0.0}, {2.5}, {2.0}, {3.0}}, []float64{1.0 / 3.0, 2.5 / 3.0}},
		{[][]float64{{0.0, 0.0}, {1.0, 1.0}, {2.0, 2.0}, {3.0, 3.0}},
			[]float64{2.0 / 3.0, 2.0 / 3.0}},
		{[][]float64{{0.0, 0.0}, {1.0, 2.0}, {2.0, 1.0}, {3.0, 4.0}},
			[]float64{(2.0/3.0 + 3.0/4.0) / 2.0, (2.0/3.0 + 2.0/4.0) / 2.0}},
	} {
		n.assignCrowdingDistance(f.in, []int{0, 1, 2, 3}, n.crowdingDistance)
		for i := 1; i < 3; i++ {
			if int(n.crowdingDistance[i]*1000000) != int(f.out[i-1]*1000000) {
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
	for testcase, f := range []struct {
		in  [][]float64
		out []int
	}{
		{[][]float64{{0.0}, {1.0}, {2.5}, {3.0}}, []int{1, 2}},
		{[][]float64{{0.0}, {0.25}, {2.5}, {3.0}}, []int{2, 1}},
	} {
		n.mixedObjectives = f.in
		n.crowdingFill(newPopulation, f.in, []int{0, 1, 2, 3}, 0)
		for i := 1; i < 3; i++ {
			if n.mixedPopulation.Individual(i).Value(0) != newPopulation.Individual(f.out[i-1]).Value(0) {
				t.Error("Expected", n.mixedPopulation.Individual(i).Value(0),
					"but was", newPopulation.Individual(f.out[i-1]).Value(0), "testcase", testcase)
			}
		}
	}
}

func TestFillNondominatedSort(t *testing.T) {
	n.previousPopulation = integer.NewRandomIntegerPopulation(4, 1, []integer.Bound{{0, 10}}, rng)
	n.merge(integer.NewRandomIntegerPopulation(4, 1, []integer.Bound{{0, 10}}, rng), nil)
	newPopulation := integer.NewRandomIntegerPopulation(4, 1, []integer.Bound{{0, 10}}, rng)
	newObjectives := [][]float64{{0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}}
	for testcase, f := range []struct {
		in   [][]float64
		out  []int
		rank []int
	}{
		{[][]float64{{0.0}, {1.0}, {2.0}, {3.0}, {4.0}, {5.0}, {6.0}, {7.0}},
			[]int{0, 1, 2, 3}, []int{1, 2, 3, 4}},
		{[][]float64{{7.0}, {6.0}, {5.0}, {4.0}, {3.0}, {2.0}, {1.0}, {0.0}},
			[]int{7, 6, 5, 4}, []int{1, 2, 3, 4}},
		{[][]float64{{0.0}, {0.0}, {0.0}, {0.0}, {0.0}, {0.0}, {0.0}, {0.0}},
			[]int{0, 7, 6, 5}, []int{1, 1, 1, 1}},
		{[][]float64{{1.0}, {0.0}, {0.0}, {1.0}, {2.0}, {2.0}, {2.0}, {2.0}},
			[]int{1, 2, 3, 0}, []int{1, 1, 2, 2}},
		{[][]float64{{0.0, 0.0}, {0.0, 0.0}, {1.5, 1.5}, {1.0, 2.0}, {2.0, 1.0},
			{5.0, 5.0}, {5.0, 5.0}, {5.0, 5.0}, {5.0, 5.0}},
			[]int{0, 1, 4, 3}, []int{1, 1, 2, 2}},
	} {
		n.mixedObjectives = f.in
		n.fillNondominatedSort(newPopulation, newObjectives)
		for i := 0; i < 4; i++ {
			if n.mixedPopulation.Individual(f.out[i]).Value(0) != newPopulation.Individual(i).Value(0) {
				t.Error("Expected", n.mixedPopulation.Individual(f.out[i]).Value(0),
					"but was", newPopulation.Individual(i).Value(0), "testcase", testcase)
			}
			if !reflect.DeepEqual(n.mixedObjectives[f.out[i]], newObjectives[i]) {
				t.Error("Expected objective", n.mixedObjectives[f.out[i]],
					"but was", newObjectives[i], "testcase", testcase)
			}
			if n.rank[i] != f.rank[i] {
				t.Error("Expected rank", f.rank[i], "but was", n.rank[i], "testcase", testcase)
			}
		}
	}
}

func TestAssignRankAndCrowdingDistance(t *testing.T) {
	for testcase, f := range []struct {
		in               [][]float64
		rank             []int
		crowdingDistance []float64
	}{
		{[][]float64{{0.0}, {1.0}, {2.0}, {3.0}}, []int{1, 2, 3, 4},
			[]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}},
		{[][]float64{{3.0}, {2.0}, {1.0}, {0.0}}, []int{4, 3, 2, 1},
			[]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}},
		{[][]float64{{0.0}, {0.0}, {0.0}, {0.0}}, []int{1, 1, 1, 1},
			[]float64{math.MaxFloat64, 0, 0, 0}},
	} {
		n.assignRankAndCrowdingDistance(f.in)
		for i, r := range f.rank {
			if n.rank[i] != r {
				t.Error("Expected rank", r, "but was", n.rank[i], "testcase", testcase)
			}
		}
		for i, d := range f.crowdingDistance {
			if n.crowdingDistance[i] != d {
				t.Error("Expected crowdingDistance", d, "but was", n.crowdingDistance[i], "testcase", testcase)
			}
		}
	}
}
