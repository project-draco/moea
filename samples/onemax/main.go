package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"project-draco.io/moea"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	fitnessFunc := func(individual moea.Individual) float64 {
		result := 0.0
		for i := 0; i < individual.Len(); i++ {
			if individual.Value(i).(bool) {
				result++
			}
		}
		return result
	}
	config := &moea.Config{
		Algorithm:            moea.NewSimpleAlgorithm(),
		Population:           moea.NewRandomBinaryPopulation(100, 20),
		FitnessFunc:          fitnessFunc,
		MaxGenerations:       50,
		CrossoverProbability: 1.0,
		MutationProbability:  0.02,
	}
	result, fitness, err := moea.Run(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(result, fitness)
}
