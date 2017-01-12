package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"time"

	"project-draco.io/moea"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	objectiveFunc := func(individual moea.Individual) float64 {
		arr := individual.Value(0).([]big.Word)
		result := 0.0
		n := 0
		for _, x := range arr {
			for ; n < individual.Len() && x != 0; x >>= 1 {
				if x&1 != 0 {
					result++
				}
				n++
			}
		}
		return result
	}
	_ /*objectiveFunc2*/ = func(individual moea.Individual) float64 {
		arr := individual.Value(0).([]bool)
		result := 0.0
		for _, x := range arr {
			if x {
				result++
			}
		}
		return result
	}
	config := &moea.Config{
		Algorithm:            moea.NewSimpleAlgorithm(10),
		Population:           moea.NewRandomBinaryPopulation(300, []int{100}),
		ObjectiveFunc:        objectiveFunc,
		MaxGenerations:       40,
		CrossoverProbability: 0.5,
		MutationProbability:  0.01,
	}
	//moea.NewRandomBooleanPopulation(300, []int{100}),
	result, objective, err := moea.Run(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(result, objective)
}
