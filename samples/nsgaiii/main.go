package main

import (
	"fmt"
	"math"
	"os"
	"time"
	"../.."
	"../../binary"
	"../../nsgaiii"
)

const (
	maxValue = float64(^uint32(0))
)

type problem struct {
	numberOfValues    int
	bounds            func(int) (float64, float64)
	objectiveFunction func(moea.Individual) []float64
}

func valueAsFloat(value interface{}, from, to float64) float64 {
	bs := value.(binary.BinaryString)
	return (to-from)*float64(bs.Int().Int64())/maxValue + from
}

var zdt6 = problem{
	10,
	func(i int) (float64, float64) { return 0, 1 },
	func(individual moea.Individual) []float64 {
		x := valueAsFloat(individual.Value(0), 0, 1)
		s := 0.0
		for i := 1; i < 10; i++ {
			s += valueAsFloat(individual.Value(i), 0, 1)
		}
		g := 1 + 9*math.Pow(s/9.0, 0.25)
		f1 := 1 - math.Exp(-4*x)*math.Pow(math.Sin(6*math.Pi*x), 6)
		return []float64{f1, g * (1 - math.Pow(f1/g, 2))}
	},
}

func main() {

	problem := zdt6

	rng := moea.NewXorshiftWithSeed(uint32(time.Now().UTC().UnixNano()))
	lengths := make([]int, problem.numberOfValues)
	for i := 0; i < problem.numberOfValues; i++ {
		lengths[i] = 32
	}
	nsgaiiiSelection := &nsgaiii.NsgaIIISelection{
		ReferencePointsDivision: 3,
	}
	config := &moea.Config{
		Algorithm:             moea.NewSimpleAlgorithm(nsgaiiiSelection, &moea.FastMutation{}),
		Population:            binary.NewRandomBinaryPopulation(100, lengths, nil, rng),
		NumberOfValues:        problem.numberOfValues,
		NumberOfObjectives:    2,
		ObjectiveFunc:         problem.objectiveFunction,
		MaxGenerations:        250,
		CrossoverProbability:  0.9,
		MutationProbability:   1.0 / (float64(problem.numberOfValues) * 32.0),
		RandomNumberGenerator: rng,
	}
	result, err := moea.Run(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for i, individual := range result.Individuals {
		fmt.Printf("[")
		for j := 0; j < config.NumberOfObjectives; j++ {
			fmt.Printf("%.4f ", individual.Objective[j])
		}
		fmt.Printf("]")
		for j := 0; j < problem.numberOfValues; j++ {
			from, to := problem.bounds(j)
			fmt.Printf(" %.2f", valueAsFloat(individual.Values[j], from, to))
		}
		fmt.Printf(" %v\n", nsgaiiiSelection.Rank[i])
	}
}
