package main

import (
	"fmt"
	"os"
	"time"

	"github.com/project-draco/moea"
	"github.com/project-draco/moea/binary"
	"github.com/project-draco/moea/nsga"
)

func main() {
	maxValue := float64(^uint32(0))
	valueAsFloat := func(value interface{}) float64 {
		bs := value.(binary.BinaryString)
		return 20*float64(bs.Int().Int64())/maxValue - 10.0
	}
	valueAsFloat3 := func(value interface{}) float64 {
		bs := value.(binary.BinaryString)
		return 40*float64(bs.Int().Int64())/maxValue - 20.0
	}
	_ /*f1*/ = func(individual moea.Individual) []float64 {
		f := valueAsFloat(individual.Value(0))
		return []float64{-f * f, -(2.0 - f) * (2.0 - f)}
	}
	_ /*f2*/ = func(individual moea.Individual) []float64 {
		a := valueAsFloat(individual.Value(0))
		x := a
		if a <= 1.0 {
			a = -a
		} else if a <= 3 {
			a = -2 + a
		} else if a <= 4 {
			a = 4 - a
		} else {
			a = -4 + a
		}
		return []float64{-a, -(x - 5) * (x - 5)}
	}
	f3 := func(individual moea.Individual) []float64 {
		a1 := valueAsFloat3(individual.Value(0))
		a2 := valueAsFloat3(individual.Value(1))
		result := []float64{
			(a1-2)*(a1-2) + (a2-1)*(a2-1) + 2,
			9*a1 - (a2-1)*(a2-1),
		}
		g := []float64{
			-(a1*a1 + a2*a2 - 225.0),
			-(a1 - 3.0*a2 + 10.0),
		}
		penalty := 0.0
		for i := 0; i < 2; i++ {
			if g[i] < 0.0 {
				penalty += 1.0e3 * g[i] * g[i]
			}
		}
		for i := 0; i < 2; i++ {
			result[i] += penalty
			result[i] = -result[i]
		}
		return result
	}

	rng := moea.NewXorshiftWithSeed(uint32(time.Now().UTC().UnixNano()))
	nsga := &nsga.NsgaSelection{
		ValuesAsFloat: func(i moea.Individual) []float64 {
			return []float64{valueAsFloat3(i.Value(0)), valueAsFloat3(i.Value(1))}
		},
		LowerBounds: []float64{ /*-10.0*/ -20.0, -20.0},
		UpperBounds: []float64{ /*10.0*/ 20.0, 20.0},
		Verbose:     true,
		Dshare:      9.0,
	}
	config := &moea.Config{
		Algorithm:             moea.NewSimpleAlgorithm(nsga, &moea.FastMutation{}),
		Population:            binary.NewRandomBinaryPopulation(100, []int{32, 32}, nil, rng),
		NumberOfValues:        2,
		NumberOfObjectives:    2,
		ObjectiveFunc:         f3,
		MaxGenerations:        500,
		CrossoverProbability:  1.0,
		MutationProbability:   0.0,
		RandomNumberGenerator: rng,
	}
	result, err := moea.Run(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, i := range result.Individuals {
		fmt.Printf("%v %.2f %.2f\n", i.Objective, valueAsFloat3(i.Values[0]), valueAsFloat3(i.Values[1]))
	}
}
