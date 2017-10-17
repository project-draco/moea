package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"project-draco.io/moea"
	"project-draco.io/moea/binary"
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
	objectiveFunc := func(individual moea.Individual) []float64 {
		bs := individual.Value(0).(binary.BinaryString)
		result := []float64{0.0}
		var w, j int
		it := bs.Iterator(&w, &j)
		l := individual.Len()
		for i := 0; i < l; i++ {
			it.Next(&w, &j)
			if it.Test(w, j) {
				result[0]--
			}
		}
		return result
	}
	_ /*objectiveFunc*/ = func(individual moea.Individual) []float64 {
		arr := individual.Value(0).([]bool)
		result := []float64{0.0}
		for _, x := range arr {
			if x {
				result[0]--
			}
		}
		return result
	}
	f := func() *moea.Config {
		rng := moea.NewXorshiftWithSeed(uint32(time.Now().UTC().UnixNano()))
		return &moea.Config{
			Algorithm: moea.NewSimpleAlgorithm(&moea.TournamentSelection{10}, &moea.FastMutation{}),
			Population: binary.NewRandomBinaryPopulation(300, []int{200},
				nil /*[]binary.Bound{{strings.Repeat("0", 200), strings.Repeat("1", 100)}}*/, rng),
			// Population:           moea.NewRandomBooleanPopulation(300, []int{200}),
			NumberOfObjectives:    1,
			ObjectiveFunc:         objectiveFunc,
			MaxGenerations:        40,
			CrossoverProbability:  0.5,
			MutationProbability:   1.0 / 200,
			RandomNumberGenerator: rng,
			OnGenerationFunc:      func(_ int, r *moea.Result) {},
		}
	}
	result, err := moea.RunRepeatedly(f, 100)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(result.BestIndividual, result.BestObjective)
}
