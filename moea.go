package moea // import "project-draco.io/moea"

type Config struct {
	Algorithm      Algorithm
	Population     Population
	MaxGenerations int
	Fitness        Fitness
}

type Algorithm interface {
	Generation(Population) (Population, error)
	Initialize(Population, Fitness) Population
}

type Population interface {
	Len() int
	Individual(int) Individual
	Fitness(int) float64
}

type Individual interface {
	Len() int
	Value(int) interface{}
	Copy(Individual, int, int) Individual
	Mutate([]bool)
}

type Fitness func(Individual) float64

func Run(config *Config) (Population, error) {
	result := config.Algorithm.Initialize(config.Population, config.Fitness)
	var err error
	for i := 0; i < config.MaxGenerations; i++ {
		result, err = config.Algorithm.Generation(result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func BestFit(population Population) int {
	var bestfit float64
	result := -1
	for i := 0; i < population.Len(); i++ {
		if i == -1 || population.Fitness(i) > bestfit {
			bestfit = population.Fitness(i)
			result = i
		}
	}
	return result
}
