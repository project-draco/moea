package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"runtime/pprof"
	"time"

	"project-draco.io/moea"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
		}()
	}
	objectiveFunc := func(individual moea.Individual) float64 {
		arr := individual.Value(0).([]big.Word)
		result := 0.0
		n := 0
		len := individual.Len()
		for _, x := range arr {
			for ; n < len && x != 0; x >>= 1 {
				if x&1 != 0 {
					result++
				}
				n++
			}
		}
		return result
	}
	_ /*objectiveFunc*/ = func(individual moea.Individual) float64 {
		arr := individual.Value(0).([]bool)
		result := 0.0
		for _, x := range arr {
			if x {
				result++
			}
		}
		return result
	}
	for i := 0; i < 100; i++ {
		config := &moea.Config{
			Algorithm:  moea.NewSimpleAlgorithm(10),
			Population: moea.NewRandomBinaryPopulation(300, []int{200}),
			// Population:           moea.NewRandomBooleanPopulation(300, []int{200}),
			ObjectiveFunc:         objectiveFunc,
			MaxGenerations:        40,
			CrossoverProbability:  0.5,
			MutationProbability:   0.01,
			RandomNumberGenerator: moea.NewXorshiftWithSeed(uint32(time.Now().UTC().UnixNano())),
		}
		_, _, err := moea.Run(config)
		// result, objective, err := moea.Run(config)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		// fmt.Println(result, objective)
	}
}
