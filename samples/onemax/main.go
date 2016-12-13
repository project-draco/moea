package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"project-draco.io/moea"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	fitness := func(individual moea.Individual) float64 {
		result := 0.0
		for i := 0; i < individual.Len(); i++ {
			if individual.Value(i).(bool) {
				result++
			}
		}
		return result
	}
	config := &moea.Config{
		Algorithm:      moea.NewSimpleAlgorithmWith(0.8, 0.02),
		Population:     moea.NewRandomBinaryPopulation(100, 20),
		MaxGenerations: 100,
		Fitness:        fitness,
	}
	result, err := moea.Run(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	bestfit := moea.BestFit(result)
	fmt.Println(result.Fitness(bestfit))
}
