package moea

import "math/rand"

type simpleAlgorithm struct {
	config *Config
}

func NewSimpleAlgorithm() Algorithm {
	return &simpleAlgorithm{}
}

func (a *simpleAlgorithm) Generation(t Population) (Population, error) {
	tt := newPopulation(t.Len())
	for i := 0; i < t.Len(); i += 2 {
		child1, child2 := a.crossover(a.selection(t), a.selection(t))
		a.mutate(child1)
		a.mutate(child2)
		tt.setIndividual(child1, i)
		tt.setIndividual(child2, i+1)
		tt.setFitness(a.config.FitnessFunc(child1), i)
		tt.setFitness(a.config.FitnessFunc(child2), i+1)
	}
	return tt, nil
}

func (a *simpleAlgorithm) selection(t Population) Individual {
	r := rand.Float64() * t.TotalFitness()
	sum := 0.0
	for i := 0; i < t.Len(); i++ {
		sum += t.Fitness(i)
		if sum >= r {
			return t.Individual(i)
		}
	}
	return t.Individual(t.Len() - 1)
}

func (a *simpleAlgorithm) crossover(parent1, parent2 Individual) (Individual, Individual) {
	if !flip(a.config.CrossoverProbability) {
		return parent1.Copy(nil, parent1.Len(), parent1.Len()),
			parent2.Copy(nil, parent2.Len(), parent2.Len())
	}
	cross := 1 + int(rand.Float64()*float64(parent1.Len()-2))
	child1 := parent1.Copy(parent2, cross, parent1.Len())
	child2 := parent2.Copy(parent1, cross, parent2.Len())
	return child1, child2
}

func (a *simpleAlgorithm) mutate(individual Individual) {
	mutations := make([]bool, individual.Len())
	for i := 0; i < individual.Len(); i++ {
		mutations[i] = flip(a.config.MutationProbability)
	}
	individual.Mutate(mutations)
}

func (a *simpleAlgorithm) Initialize(config *Config) Population {
	a.config = config
	pp := newPopulation(config.Population.Len())
	for i := 0; i < config.Population.Len(); i++ {
		pp.setIndividual(config.Population.Individual(i), i)
		pp.setFitness(a.config.FitnessFunc(pp.Individual(i)), i)
	}
	return pp
}
