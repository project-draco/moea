package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/JoaoGabriel0511/moea"
	"github.com/JoaoGabriel0511/moea/binary"
	"github.com/JoaoGabriel0511/moea/nsgaii"
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

var fon = problem{
	3,
	func(i int) (float64, float64) { return -4, 4 },
	func(individual moea.Individual) []float64 {
		s1, s2 := 0.0, 0.0
		for i := 0; i < 3; i++ {
			x := valueAsFloat(individual.Value(i), -4, 4)
			s1 += math.Pow(x-1/math.Sqrt(3), 2)
			s2 += math.Pow(x+1/math.Sqrt(3), 2)
		}
		return []float64{1 - math.Exp(-s1), 1 - math.Exp(-s2)}
	},
}

var pol = problem{
	2,
	func(i int) (float64, float64) { return -math.Pi, math.Pi },
	func(individual moea.Individual) []float64 {
		x1 := valueAsFloat(individual.Value(0), -math.Pi, math.Pi)
		x2 := valueAsFloat(individual.Value(1), -math.Pi, math.Pi)
		a1 := 0.5*math.Sin(1) - 2*math.Cos(1) + math.Sin(2) - 1.5*math.Cos(2)
		a2 := 1.5*math.Sin(1) - math.Cos(1) + 2*math.Sin(2) - 0.5*math.Cos(2)
		b1 := 0.5*math.Sin(x1) - 2*math.Cos(x1) + math.Sin(x2) - 1.5*math.Cos(x2)
		b2 := 1.5*math.Sin(x1) - math.Cos(x1) + 2*math.Sin(x2) - 0.5*math.Cos(x2)
		return []float64{1 + math.Pow(a1-b1, 2) + math.Pow(a2-b2, 2),
			math.Pow(x1+3, 2) + math.Pow(x2+1, 2)}
	},
}

var kur = problem{
	3,
	func(i int) (float64, float64) { return -5, 5 },
	func(individual moea.Individual) []float64 {
		s1, s2 := 0.0, 0.0
		for i := 0; i < 3; i++ {
			x := valueAsFloat(individual.Value(i), -5, 5)
			if i < 2 {
				xx := valueAsFloat(individual.Value(i+1), -5, 5)
				s1 += -10 * math.Exp(-0.2*math.Sqrt(x*x+xx*xx))
			}
			s2 += math.Pow(math.Abs(x), 0.8) + 5*math.Sin(x*x*x)
		}
		return []float64{s1, s2}
	},
}

var zdt1 = problem{
	30,
	func(i int) (float64, float64) { return 0, 1 },
	func(individual moea.Individual) []float64 {
		x := valueAsFloat(individual.Value(0), 0, 1)
		s := 0.0
		for i := 1; i < 30; i++ {
			s += valueAsFloat(individual.Value(i), 0, 1)
		}
		g := (1 + 9*s/29)
		return []float64{x, g * (1 - math.Sqrt(x/g))}
	},
}

var zdt2 = problem{
	30,
	func(i int) (float64, float64) { return 0, 1 },
	func(individual moea.Individual) []float64 {
		x := valueAsFloat(individual.Value(0), 0, 1)
		s := 0.0
		for i := 1; i < 30; i++ {
			s += valueAsFloat(individual.Value(i), 0, 1)
		}
		g := (1 + 9*s/29)
		return []float64{x, g * (1 - math.Pow(x/g, 2))}
	},
}

var zdt3 = problem{
	30,
	func(i int) (float64, float64) { return 0, 1 },
	func(individual moea.Individual) []float64 {
		x := valueAsFloat(individual.Value(0), 0, 1)
		s := 0.0
		for i := 1; i < 30; i++ {
			s += valueAsFloat(individual.Value(i), 0, 1)
		}
		g := (1 + 9*s/29)
		return []float64{x, g * (1 - math.Sqrt(x/g) - x/g*math.Sin(10*math.Pi*x))}
	},
}

var zdt4 = problem{
	10,
	func(i int) (float64, float64) {
		if i == 0 {
			return 0, 1
		} else {
			return -5, 5
		}
	},
	func(individual moea.Individual) []float64 {
		x := valueAsFloat(individual.Value(0), 0, 1)
		g := 0.0
		for i := 1; i < 10; i++ {
			xx := valueAsFloat(individual.Value(i), -5, 5)
			g += xx*xx - 10.0*math.Cos(4.0*math.Pi*xx)
		}
		g += 91.0
		return []float64{x, g * (1.0 - math.Sqrt(x/g))}
	},
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

func valueAsFloat(value interface{}, from, to float64) float64 {
	bs := value.(binary.BinaryString)
	return (to-from)*float64(bs.Int().Int64())/maxValue + from
}

func main() {

	problem := zdt6

	rng := moea.NewXorshiftWithSeed(uint32(time.Now().UTC().UnixNano()))
	lengths := make([]int, problem.numberOfValues)
	for i := 0; i < problem.numberOfValues; i++ {
		lengths[i] = 32
	}
	nsgaiiSelection := &nsgaii.NsgaIISelection{
		NsgaiiiVariant: nil,
	}
	config := &moea.Config{
		Algorithm:             moea.NewSimpleAlgorithm(nsgaiiSelection, &moea.FastMutation{}),
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
		fmt.Printf(" %v\n", nsgaiiSelection.Rank[i])
	}
}
