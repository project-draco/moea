package moea

import "math/rand"

type simpleAlgorithm struct {
	config        *Config
	oldObjectives map[int]float64
	newObjectives map[int]float64
	objectivesSum float64
	oldPopulation Population
	newPopulation Population
	mutations     []bool
}

func NewSimpleAlgorithm() Algorithm {
	return &simpleAlgorithm{}
}

func (a *simpleAlgorithm) Generation() (Individual, float64, error) {
	var bestIndividual Individual
	var bestFit float64
	newObjectivesSum := 0.0
	for i := 0; i < a.newPopulation.Len(); i += 2 {
		child1 := a.newPopulation.Individual(i)
		child2 := a.newPopulation.Individual(i + 1)
		a.crossover(a.selection(), a.selection(), child1, child2)
		a.mutate(child1)
		a.mutate(child2)
		f1 := a.config.ObjectiveFunc(child1)
		f2 := a.config.ObjectiveFunc(child2)
		a.newObjectives[i] = f1
		a.newObjectives[i+1] = f2
		newObjectivesSum += f1 + f2
		if f1 > bestFit {
			bestFit = f1
			bestIndividual = child1
		}
		if f2 > bestFit {
			bestFit = f2
			bestIndividual = child2
		}
	}
	a.oldObjectives, a.newObjectives = a.newObjectives, a.oldObjectives
	a.objectivesSum = newObjectivesSum
	a.oldPopulation, a.newPopulation = a.newPopulation, a.oldPopulation
	return bestIndividual, bestFit, nil
}

func (a *simpleAlgorithm) selection() Individual {
	r := rand.Float64() * a.objectivesSum
	sum := 0.0
	for i := 0; i < a.oldPopulation.Len(); i++ {
		sum += a.oldObjectives[i]
		if sum >= r {
			return a.oldPopulation.Individual(i)
		}
	}
	return a.oldPopulation.Individual(a.oldPopulation.Len() - 1)
}

func (a *simpleAlgorithm) crossover(parent1, parent2, child1, child2 Individual) {
	if !flip(a.config.CrossoverProbability) {
		child1.Copy(parent1, 0, child1.Len())
		child2.Copy(parent2, 0, child2.Len())
		return
	}
	cross := 1 + int(rand.Float64()*float64(parent1.Len()-2))
	child1.Copy(parent1, 0, cross)
	child1.Copy(parent2, cross, child1.Len())
	child2.Copy(parent2, 0, cross)
	child2.Copy(parent1, cross, child2.Len())
}

func (a *simpleAlgorithm) mutate(individual Individual) {
	for i := 0; i < individual.Len(); i++ {
		a.mutations[i] = flip(a.config.MutationProbability)
	}
	individual.Mutate(a.mutations)
}

func (a *simpleAlgorithm) Initialize(config *Config) {
	a.config = config
	a.newObjectives = map[int]float64{}
	a.oldObjectives = map[int]float64{}
	for i := 0; i < config.Population.Len(); i++ {
		a.oldObjectives[i] = a.config.ObjectiveFunc(config.Population.Individual(i))
		a.objectivesSum += a.oldObjectives[i]
	}
	a.oldPopulation = config.Population
	a.newPopulation = config.Population.Clone()
	a.mutations = make([]bool, a.oldPopulation.Individual(0).Len())
}
