package moea // import "project-draco.io/moea"

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
