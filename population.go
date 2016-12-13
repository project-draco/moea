package moea

type population struct {
	arr       []Individual
	fitnesses map[int]float64
}

func newPopulation(size int) *population {
	return &population{arr: make([]Individual, size), fitnesses: make(map[int]float64, 0)}
}

func (p *population) Len() int { return len(p.arr) }

func (p *population) Individual(i int) Individual { return p.arr[i] }

func (p *population) Fitness(i int) float64 {
	return p.fitnesses[i]
}

func (p *population) setIndividual(individual Individual, i int) {
	p.arr[i] = individual
}

func (p *population) setFitness(value float64, i int) {
	p.fitnesses[i] = value
}
