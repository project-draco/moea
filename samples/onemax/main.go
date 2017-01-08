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
	config := &moea.Config{
		Algorithm:            moea.NewSimpleAlgorithm(),
		Population:           moea.NewRandomBinaryPopulation(100, []int{20}),
		ObjectiveFunc:        objectiveFunc,
		MaxGenerations:       50,
		CrossoverProbability: 1.0,
		MutationProbability:  0.02,
	}
	result, objective, err := moea.Run(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(result, objective)
}
