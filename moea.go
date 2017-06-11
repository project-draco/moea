package moea // import "project-draco.io/moea"
import "runtime"

type Config struct {
	Algorithm             Algorithm
	Population            Population
	ObjectiveFunc         ObjectiveFunc
	MaxGenerations        int
	CrossoverProbability  float64
	MutationProbability   float64
	RandomNumberGenerator RNG
	OnGenerationFunc      OnGenerationFunc
}

type Algorithm interface {
	Generation() (*Result, error)
	Initialize(*Config)
}

type Population interface {
	Len() int
	Individual(int) Individual
	Clone() Population
}

type Individual interface {
	Len() int
	Value(int) interface{}
	Copy(Individual, int, int)
	Mutate([]int)
	Clone() Individual
}

type ObjectiveFunc func(Individual) float64

type OnGenerationFunc func(int, *Result)

type Result struct {
	BestIndividual   Individual
	BestObjective    float64
	WorstObjective   float64
	AverageObjective float64
	Mutations        int
	Crossovers       int
	Individuals      []IndividualResult
}

type IndividualResult struct {
	Objective float64
	Parent1   int
	Parent2   int
	CrossSite int
}

func Run(config *Config) (*Result, error) {
	result := &Result{}
	config.Algorithm.Initialize(config)
	result.BestIndividual = config.Population.Individual(0).Clone()
	for i := 0; i < config.MaxGenerations; i++ {
		generationResult, err := config.Algorithm.Generation()
		if err != nil {
			return nil, err
		}
		if config.OnGenerationFunc != nil {
			config.OnGenerationFunc(i, generationResult)
		}
		if generationResult.BestObjective > result.BestObjective {
			result.BestIndividual.Copy(generationResult.BestIndividual, 0, result.BestIndividual.Len())
			result.BestObjective = generationResult.BestObjective
		}
		result.Mutations += generationResult.Mutations
		result.Crossovers += generationResult.Crossovers
	}
	return result, nil
}

func RunRepeatedly(configfunc func() *Config, repeat int) (*Result, error) {
	if repeat < 2 {
		return Run(configfunc())
	}
	var numCPU = runtime.GOMAXPROCS(0)
	bestResults := make([]*Result, numCPU)
	c := make(chan int, numCPU)
	for i := 0; i < numCPU; i++ {
		cpu := i
		go func() {
			for j := 0; j < repeat/numCPU; j++ {
				result, err := Run(configfunc())
				if err != nil {
					panic(err)
				}
				if bestResults[cpu] == nil || bestResults[cpu].BestObjective < result.BestObjective {
					bestResults[cpu] = result
				}
			}
			c <- 1
		}()
	}
	for i := 0; i < numCPU; i++ {
		<-c
	}
	var bestResult *Result
	for i := 0; i < numCPU; i++ {
		if bestResult == nil || bestResult.BestObjective < bestResults[i].BestObjective {
			bestResult = bestResults[i]
		}
	}
	return bestResult, nil
}
