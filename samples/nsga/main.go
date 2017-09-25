package main

import (
	"fmt"
	"os"
	"time"

	"project-draco.io/moea"
	"project-draco.io/moea/binary"
	"project-draco.io/moea/nsga"
)

func main() {
	maxValue := float64(^uint32(0))
	valueAsFloat := func(value interface{}) float64 {
		bs := value.(binary.BinaryString)
		return 20*float64(bs.Int().Int64())/maxValue - 10.0
	}
	f1 := func(individual moea.Individual) []float64 {
		f := valueAsFloat(individual.Value(0))
		return []float64{-f * f, -(2.0 - f) * (2.0 - f)}
	}
	rng := moea.NewXorshiftWithSeed(uint32(time.Now().UTC().UnixNano()))
	nsga := &nsga.NsgaSelection{
		ValuesAsFloat: func(i moea.Individual) []float64 {
			return []float64{valueAsFloat(i.Value(0))}
		},
		LowerBounds: []float64{-10.0},
		UpperBounds: []float64{10.0},
	}
	config := &moea.Config{
		Algorithm:             moea.NewSimpleAlgorithm(nsga),
		Population:            binary.NewRandomBinaryPopulation(100, []int{32}, nil, rng),
		NumberOfValues:        1,
		NumberOfObjectives:    2,
		ObjectiveFunc:         f1,
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
		fmt.Println(i.Objective, valueAsFloat(i.Values[0]))
	}
}
