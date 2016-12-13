package moea

import "math/rand"

type simpleAlgorithm struct {
	fitness              Fitness
	fitnessSum           float64
	crossoverProbability float64
	mutationProbability  float64
}

func NewSimpleAlgorithm() Algorithm {
	return NewSimpleAlgorithmWith(0.6, 0.0333)
}

func NewSimpleAlgorithmWith(crossoverProbability, mutationProbability float64) Algorithm {
	return &simpleAlgorithm{
		crossoverProbability: crossoverProbability,
		mutationProbability:  mutationProbability,
	}
}

func (a *simpleAlgorithm) Generation(t Population) (Population, error) {
	tt := newPopulation(t.Len())
	newSum := 0.0
	for i := 0; i < t.Len(); i += 2 {
		child1, child2 := a.crossover(a.selection(t), a.selection(t))
		a.mutate(child1)
		a.mutate(child2)
		tt.setIndividual(child1, i)
		tt.setIndividual(child2, i+1)
		tt.setFitness(a.fitness(child1), i)
		tt.setFitness(a.fitness(child2), i+1)
		newSum += tt.Fitness(i) + tt.Fitness(i+1)
	}
	a.fitnessSum = newSum
	return tt, nil
}

func (a *simpleAlgorithm) selection(t Population) Individual {
	r := rand.Float64() * a.fitnessSum
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
	if !flip(a.crossoverProbability) {
		return parent1.Copy(nil, parent1.Len(), parent1.Len()),
			parent2.Copy(nil, parent2.Len(), parent2.Len())
	}
	cross := int(rand.Float64() * float64(parent1.Len()))
	child1 := parent1.Copy(parent2, cross, parent1.Len())
	child2 := parent2.Copy(parent1, cross, parent2.Len())
	return child1, child2
}

func (a *simpleAlgorithm) mutate(individual Individual) {
	mutations := make([]bool, individual.Len())
	for i := 0; i < individual.Len(); i++ {
		mutations[i] = flip(a.mutationProbability)
	}
	individual.Mutate(mutations)
}

func (a *simpleAlgorithm) Initialize(p Population, fitness Fitness) Population {
	a.fitness = fitness
	pp := newPopulation(p.Len())
	sum := 0.0
	for i := 0; i < p.Len(); i++ {
		pp.setIndividual(p.Individual(i), i)
		f := fitness(pp.Individual(i))
		pp.setFitness(f, i)
		sum += f
	}
	a.fitnessSum = sum
	return pp
}
