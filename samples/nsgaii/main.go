package main

import (
	"fmt"
	"os"
	"time"

	"project-draco.io/moea"
	"project-draco.io/moea/binary"
	"project-draco.io/moea/nsgaii"
)

const (
	maxValue = float64(^uint32(0))
)

type problem struct {
	numberOfValues    int
	bounds            func(int) (float64, float64)
	objectiveFunction func(moea.Individual) []float64
}

var sch = problem{
	1,
	func(i int) (float64, float64) { return -1000, 1000 },
	func(individual moea.Individual) []float64 {
		x := valueAsFloat(individual.Value(0), -1000, 1000)
		return []float64{x * x, (x - 2.0) * (x - 2.0)}
	},
}

func valueAsFloat(value interface{}, from, to float64) float64 {
	bs := value.(binary.BinaryString)
	return (to-from)*float64(bs.Int().Int64())/maxValue + from
}

func main() {

	problem := sch

	rng := moea.NewXorshiftWithSeed(uint32(time.Now().UTC().UnixNano()))
	config := &moea.Config{
		Algorithm:             moea.NewSimpleAlgorithm(&nsgaii.NsgaIISelection{}),
		Population:            binary.NewRandomBinaryPopulation(100, []int{32, 32}, nil, rng),
		NumberOfValues:        problem.numberOfValues,
		NumberOfObjectives:    2,
		ObjectiveFunc:         problem.objectiveFunction,
		MaxGenerations:        250,
		CrossoverProbability:  0.9,
		MutationProbability:   1.0 / 64.0,
		RandomNumberGenerator: rng,
	}
	result, err := moea.Run(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, i := range result.Individuals {
		fmt.Printf("%v", i.Objective)
		for j := 0; j < problem.numberOfValues; j++ {
			from, to := problem.bounds(j)
			fmt.Printf(" %.2f", valueAsFloat(i.Values[j], from, to))
		}
		fmt.Println()
	}
}
