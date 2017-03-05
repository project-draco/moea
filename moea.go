package moea // import "project-draco.io/moea"

type Config struct {
	Algorithm             Algorithm
	Population            Population
	ObjectiveFunc         ObjectiveFunc
	MaxGenerations        int
	CrossoverProbability  float64
	MutationProbability   float64
	RandomNumberGenerator RNG
}

type Algorithm interface {
	Generation() (Individual, float64, error)
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
	Mutate([]bool)
	Clone() Individual
}

type ObjectiveFunc func(Individual) float64

func Run(config *Config) (Individual, float64, error) {
	config.Algorithm.Initialize(config)
	bestIndividualEver := config.Population.Individual(0).Clone()
	bestObjectiveEver := 0.0
	for i := 0; i < config.MaxGenerations; i++ {
		individual, objective, err := config.Algorithm.Generation()
		if err != nil {
			return nil, 0.0, err
		}
		if objective > bestObjectiveEver {
			bestIndividualEver.Copy(individual, 0, bestIndividualEver.Len())
			bestObjectiveEver = objective
		}
	}
	return bestIndividualEver, bestObjectiveEver, nil
}
