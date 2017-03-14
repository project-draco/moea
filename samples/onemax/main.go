package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
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
		bs := individual.Value(0).(moea.BinaryString)
		result := 0.0
		for i := bs.Iterator(); i.Next(); {
			if i.Test() {
				result++
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
	f := func(seed uint32) {
		rng := moea.NewXorshiftWithSeed(seed)
		config := &moea.Config{
			Algorithm:  moea.NewSimpleAlgorithm(10),
			Population: moea.NewRandomBinaryPopulation(300, []int{200}, nil, rng),
			// Population:           moea.NewRandomBooleanPopulation(300, []int{200}),
			ObjectiveFunc:         objectiveFunc,
			MaxGenerations:        40,
			CrossoverProbability:  0.5,
			MutationProbability:   0.01,
			RandomNumberGenerator: rng,
		}
		_, _, err := moea.Run(config)
		// result, objective, err := moea.Run(config)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		// fmt.Println(result, objective)
	}
	var numCPU = runtime.GOMAXPROCS(0)
	c := make(chan int, numCPU)
	for i := 0; i < numCPU; i++ {
		go func() {
			for j := 0; j < 100/numCPU; j++ {
				f(uint32(time.Now().UTC().UnixNano()))
			}
			c <- 1
		}()
	}
	for i := 0; i < numCPU; i++ {
		<-c
	}
}
