package moea // import "project-draco.io/moea"

type Config struct {
	Algorithm      Algorithm
	Population     Population
	FitnessFunc    FitnessFunc
	MaxGenerations int
}

type Algorithm interface {
	Generation(Population) (Population, error)
	Initialize(Population, FitnessFunc) Population
}

type Population interface {
	Len() int
	Individual(int) Individual
	Fitness(int) float64
	TotalFitness() float64
}

type Individual interface {
	Len() int
	Value(int) interface{}
	Copy(Individual, int, int) Individual
	Mutate([]bool)
}

type FitnessFunc func(Individual) float64

func Run(config *Config) (Individual, float64, error) {
	population := config.Algorithm.Initialize(config.Population, config.FitnessFunc)
	var result Individual
	var err error
	bestfit := 0.0
	for i := 0; i < config.MaxGenerations; i++ {
		population, err = config.Algorithm.Generation(population)
		if err != nil {
			return nil, 0.0, err
		}
		for j := 0; j < population.Len(); j++ {
			if population.Fitness(j) > bestfit {
				result = population.Individual(j)
				bestfit = population.Fitness(j)
			}
		}
	}
	return result, bestfit, nil
}
