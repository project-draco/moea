package moea

import (
	"math"
	//"fmt"
)

type simpleAlgorithm struct {
	config               *Config
	oldObjectives        [][]float64
	newObjectives        [][]float64
	objectivesSum        []float64
	oldPopulation        Population
	newPopulation        Population
	selectionOperator    SelectionOperator
	mutationOperator     MutationOperator
	crossoverProbability float64
	mutationProbability  float64
	result               *Result
}

type SelectionOperator interface {
	Selection(config *Config, objectives [][]float64) int
}

type MutationOperator interface {
	Mutation(config *Config, individual Individual, probability float64)
}

type RouletteWheelSelection struct{ objectivesSum float64 }

type TournamentSelection struct{ TournamentSize int }

type RegularMutation struct {
	mutationsIndexes []int
	mutations        int
}

type FastMutation struct{ RegularMutation }

func NewSimpleAlgorithm(selectionOperator SelectionOperator, mutationOperator MutationOperator) Algorithm {
	if selectionOperator == nil {
		selectionOperator = &TournamentSelection{10}
	}
	if mutationOperator == nil {
		mutationOperator = &RegularMutation{}
	}
	a := &simpleAlgorithm{selectionOperator: selectionOperator, mutationOperator: mutationOperator}
	return a
}

func (a *simpleAlgorithm) Generation() (*Result, error) {
	*a.result = Result{
		Individuals:      a.result.Individuals,
		AverageObjective: a.result.AverageObjective,
		WorstObjective:   a.result.WorstObjective,
		BestObjective:    a.result.BestObjective,
	}
	for i := 0; i < a.config.NumberOfObjectives; i++ {
		a.objectivesSum[i] = 0
		a.result.AverageObjective[i] = 0
		a.result.WorstObjective[i] = 0
		a.result.BestObjective[i] = math.MaxFloat64
	}
	type onGenerationListener interface {
		OnGeneration(*Config, Population, [][]float64)
	}
	if l, ok := a.selectionOperator.(onGenerationListener); ok {
		//fmt.Printf("ahhhh\n")
		l.OnGeneration(a.config, a.oldPopulation, a.oldObjectives)
		//fmt.Printf("ihhhh\n")
	}
	for i := 0; i < a.newPopulation.Len(); i += 2 {
		child1 := a.newPopulation.Individual(i)
		child2 := a.newPopulation.Individual(i + 1)
		parentIndex1 := a.selectionOperator.Selection(a.config, a.oldObjectives)
		parentIndex2 := a.selectionOperator.Selection(a.config, a.oldObjectives)
		parent1 := a.oldPopulation.Individual(parentIndex1)
		parent2 := a.oldPopulation.Individual(parentIndex2)
		crossSite := a.crossover(parent1, parent2, child1, child2)
		a.mutationOperator.Mutation(a.config, child1, a.mutationProbability)
		a.mutationOperator.Mutation(a.config, child2, a.mutationProbability)
		f1 := a.config.ObjectiveFunc(child1)
		f2 := a.config.ObjectiveFunc(child2)
		a.newObjectives[i] = f1
		a.newObjectives[i+1] = f2
		if f1[0] <= f2[0] && f1[0] < a.result.BestObjective[0] {
			a.result.BestIndividual = child1
		} else if f2[0] <= f1[0] && f2[0] < a.result.BestObjective[0] {
			a.result.BestIndividual = child2
		}
		for j := 0; j < a.config.NumberOfObjectives; j++ {
			a.objectivesSum[j] += f1[j] + f2[j]
			if math.Min(f1[j], f2[j]) < a.result.BestObjective[j] {
				a.result.BestObjective[j] = math.Min(f1[j], f2[j])
			}
			if math.Max(f1[j], f2[j]) > a.result.WorstObjective[j] {
				a.result.WorstObjective[j] = math.Max(f1[j], f2[j])
			}
		}
		a.result.Individuals[i].Objective = f1
		a.result.Individuals[i+1].Objective = f2
		a.result.Individuals[i].Parent1 = parentIndex1
		a.result.Individuals[i].Parent2 = parentIndex2
		a.result.Individuals[i+1].Parent1 = parentIndex1
		a.result.Individuals[i+1].Parent2 = parentIndex2
		a.result.Individuals[i].CrossSite = crossSite
		a.result.Individuals[i+1].CrossSite = crossSite
		if a.result.Individuals[i].Values == nil {
			a.result.Individuals[i].Values = make([]interface{}, a.config.NumberOfValues)
			a.result.Individuals[i+1].Values = make([]interface{}, a.config.NumberOfValues)
		}
		for j := 0; j < a.config.NumberOfValues; j++ {
			a.result.Individuals[i].Values[j] = child1.Value(j)
			a.result.Individuals[i+1].Values[j] = child2.Value(j)
		}
	}
	a.oldObjectives, a.newObjectives = a.newObjectives, a.oldObjectives
	a.oldPopulation, a.newPopulation = a.newPopulation, a.oldPopulation
	for i := 0; i < a.config.NumberOfObjectives; i++ {
		a.result.AverageObjective[i] = a.objectivesSum[i] / float64(a.newPopulation.Len())
	}
	return a.result, nil
}

func (a *simpleAlgorithm) Finalize(result *Result) {
	type finalizer interface {
		Finalize(*Config, Population, [][]float64, *Result)
	}
	if f, ok := a.selectionOperator.(finalizer); ok {
		f.Finalize(a.config, a.oldPopulation, a.oldObjectives, result)
	}
	if f, ok := a.mutationOperator.(finalizer); ok {
		f.Finalize(a.config, a.oldPopulation, a.oldObjectives, result)
	}
}

func (rws *RouletteWheelSelection) OnGeneration(config *Config, objectives [][]float64) {
	rws.objectivesSum = 0
	for _, o := range objectives {
		rws.objectivesSum += o[0]
	}
}

func (rws *RouletteWheelSelection) Selection(config *Config, objectives [][]float64) int {
	r := config.RandomNumberGenerator.Float64() * rws.objectivesSum
	sum := 0.0
	for i := 0; i < config.Population.Len(); i++ {
		sum += objectives[i][0]
		if -sum >= -r {
			return i
		}
	}
	return config.Population.Len() - 1
}

func (ts *TournamentSelection) Selection(config *Config, objectives [][]float64) int {
	result := -1
	for i := 0; i < ts.TournamentSize; i++ {
		r := int(config.RandomNumberGenerator.Float64() * float64(config.Population.Len()-1))
		if result == -1 || objectives[r][0] < objectives[result][0] {
			result = r
		}
	}
	return result
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

func (m *RegularMutation) Initialize(config *Config) {
	m.mutationsIndexes = make([]int, config.Population.Individual(0).Len())
}

func (m *RegularMutation) Mutation(config *Config, individual Individual, probability float64) {
	len := individual.Len()
	j := 0
	for i := 0; i < len; i++ {
		f := config.RandomNumberGenerator.Flip(probability)
		if f {
			m.mutationsIndexes[j] = i
			j++
			m.mutations++
		}
	}
	individual.Mutate(m.mutationsIndexes[0:j])
}

func (m *RegularMutation) Finalize(_ *Config, _ Population, _ [][]float64, result *Result) {
	result.Mutations = m.mutations
}

func (m *FastMutation) Initialize(config *Config) {
	m.RegularMutation.Initialize(config)
}

func (m *FastMutation) Mutation(config *Config, individual Individual, probability float64) {
	j := 0
	for len := float64(individual.Len()) * probability; len > 0; len-- {
		i := int(config.RandomNumberGenerator.Float64() * float64(individual.Len()))
		if i == individual.Len() {
			i = individual.Len() - 1
		}
		if len >= 1 || config.RandomNumberGenerator.Flip(len) {
			m.mutationsIndexes[j] = i
			j++
			m.mutations++
		}
	}
	individual.Mutate(m.mutationsIndexes[0:j])
}

func (m *FastMutation) Finalize(_ *Config, _ Population, _ [][]float64, result *Result) {
	m.RegularMutation.Finalize(nil, nil, nil, result)
}

func (a *simpleAlgorithm) Initialize(config *Config) {
	a.config = config
	a.oldObjectives = make([][]float64, config.Population.Len())
	a.newObjectives = make([][]float64, config.Population.Len())
	a.objectivesSum = make([]float64, config.NumberOfObjectives)
	for i := 0; i < config.Population.Len(); i++ {
		a.oldObjectives[i] = a.config.ObjectiveFunc(config.Population.Individual(i))
		for j := 0; j < config.NumberOfObjectives; j++ {
			a.objectivesSum[j] += a.oldObjectives[i][j]
		}
	}
	a.oldPopulation = config.Population
	a.newPopulation = config.Population.Clone()
	a.crossoverProbability = a.config.CrossoverProbability
	a.mutationProbability = a.config.MutationProbability
	a.result = &Result{
		Individuals:      make([]IndividualResult, config.Population.Len()),
		AverageObjective: make([]float64, a.config.NumberOfObjectives),
		WorstObjective:   make([]float64, a.config.NumberOfObjectives),
		BestObjective:    make([]float64, a.config.NumberOfObjectives),
	}
	type initializer interface {
		Initialize(*Config)
	}
	if i, ok := a.selectionOperator.(initializer); ok {
		i.Initialize(a.config)
	}
	if i, ok := a.mutationOperator.(initializer); ok {
		i.Initialize(a.config)
	}
}
