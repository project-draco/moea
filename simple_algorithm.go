package moea

import (
	"math"
)

type simpleAlgorithm struct {
	config               *Config
	oldObjectives        []float64
	newObjectives        []float64
	objectivesSum        float64
	oldPopulation        Population
	newPopulation        Population
	mutations            []bool
	mutationsIndexes     []int
	tournamentSize       int
	crossoverProbability float64
	mutationProbability  float64
	result               *Result
}

func NewSimpleAlgorithm(tournamentSize int) Algorithm {
	return &simpleAlgorithm{tournamentSize: tournamentSize}
}

func (a *simpleAlgorithm) Generation() (*Result, error) {
	*a.result = Result{Individuals: a.result.Individuals}
	newObjectivesSum := 0.0
	for i := 0; i < a.newPopulation.Len(); i += 2 {
		child1 := a.newPopulation.Individual(i)
		child2 := a.newPopulation.Individual(i + 1)
		parent1, parentIndex1 := a.tournamentSelection()
		parent2, parentIndex2 := a.tournamentSelection()
		crossSite := a.crossover(parent1, parent2, child1, child2)
		a.mutate(child1)
		a.mutate(child2)
		f1 := a.config.ObjectiveFunc(child1)
		f2 := a.config.ObjectiveFunc(child2)
		a.newObjectives[i] = f1
		a.newObjectives[i+1] = f2
		newObjectivesSum += f1 + f2
		if math.Max(f1, f2) > a.result.BestObjective {
			a.result.BestObjective = math.Max(f1, f2)
			if f1 > f2 {
				a.result.BestIndividual = child1
			} else {
				a.result.BestIndividual = child2
			}
		}
		if math.Min(f1, f2) < a.result.WorstObjective {
			a.result.WorstObjective = math.Min(f1, f2)
		}
		a.result.Individuals[i].Objective = f1
		a.result.Individuals[i+1].Objective = f2
		a.result.Individuals[i].Parent1 = parentIndex1
		a.result.Individuals[i].Parent2 = parentIndex2
		a.result.Individuals[i+1].Parent1 = parentIndex1
		a.result.Individuals[i+1].Parent2 = parentIndex2
		a.result.Individuals[i].CrossSite = crossSite
		a.result.Individuals[i+1].CrossSite = crossSite
	}
	a.oldObjectives, a.newObjectives = a.newObjectives, a.oldObjectives
	a.objectivesSum = newObjectivesSum
	a.oldPopulation, a.newPopulation = a.newPopulation, a.oldPopulation
	a.result.AverageObjective = newObjectivesSum / float64(a.newPopulation.Len())
	return a.result, nil
}

func (a *simpleAlgorithm) rouletteWheelSelection() Individual {
	r := a.config.RandomNumberGenerator.Float64() * a.objectivesSum
	sum := 0.0
	for i := 0; i < a.oldPopulation.Len(); i++ {
		sum += a.oldObjectives[i]
		if sum >= r {
			return a.oldPopulation.Individual(i)
		}
	}
	return a.oldPopulation.Individual(a.oldPopulation.Len() - 1)
}

func (a *simpleAlgorithm) tournamentSelection() (Individual, int) {
	result := -1
	for i := 0; i < a.tournamentSize; i++ {
		r := int(a.config.RandomNumberGenerator.Float64() * float64(a.oldPopulation.Len()-1))
		if result == -1 || a.oldObjectives[r] > a.oldObjectives[result] {
			result = r
		}
	}
	return a.oldPopulation.Individual(result), result
}

func (a *simpleAlgorithm) crossover(parent1, parent2, child1, child2 Individual) int {
	if !a.config.RandomNumberGenerator.Flip(a.crossoverProbability) {
		child1.Copy(parent1, 0, child1.Len())
		child2.Copy(parent2, 0, child2.Len())
		return -1
	}
	cross := 1 + int(a.config.RandomNumberGenerator.Float64()*float64(parent1.Len()-2))
	child1.Copy(parent1, 0, cross)
	child1.Copy(parent2, cross, child1.Len())
	child2.Copy(parent2, 0, cross)
	child2.Copy(parent1, cross, child2.Len())
	a.result.Crossovers++
	return cross
}

func (a *simpleAlgorithm) mutate(individual Individual) {
	len := individual.Len()
	j := 0
	for i := 0; i < len; i++ {
		f := a.config.RandomNumberGenerator.Flip(a.mutationProbability)
		if f {
			a.mutationsIndexes[j] = i
			j++
			a.result.Mutations++
		}
	}
	individual.Mutate(a.mutationsIndexes[0:j])
}

func (a *simpleAlgorithm) Initialize(config *Config) {
	a.config = config
	a.oldObjectives = make([]float64, config.Population.Len())
	a.newObjectives = make([]float64, config.Population.Len())
	for i := 0; i < config.Population.Len(); i++ {
		a.oldObjectives[i] = a.config.ObjectiveFunc(config.Population.Individual(i))
		a.objectivesSum += a.oldObjectives[i]
	}
	a.oldPopulation = config.Population
	a.newPopulation = config.Population.Clone()
	a.mutations = make([]bool, a.oldPopulation.Individual(0).Len())
	a.mutationsIndexes = make([]int, a.oldPopulation.Individual(0).Len())
	a.crossoverProbability = a.config.CrossoverProbability * float64(MaxUint32)
	a.mutationProbability = a.config.MutationProbability * float64(MaxUint32)
	a.result = &Result{Individuals: make([]IndividualResult, config.Population.Len())}
}
