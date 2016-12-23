package moea

type population struct {
	arr          []Individual
	fitnesses    map[int]float64
	totalFitness float64
}

func newPopulation(size int) *population {
	return newPopulationWith(make([]Individual, size), size)
}

func newPopulationWith(arr []Individual, size int) *population {
	return &population{arr: arr, fitnesses: make(map[int]float64, 0)}
}

func (p *population) Len() int { return len(p.arr) }

func (p *population) Individual(i int) Individual { return p.arr[i] }

func (p *population) Fitness(i int) float64 {
	return p.fitnesses[i]
}

func (p *population) TotalFitness() float64 { return p.totalFitness }

func (p *population) setIndividual(individual Individual, i int) {
	p.arr[i] = individual
}

func (p *population) setFitness(value float64, i int) {
	oldValue := p.fitnesses[i]
	p.fitnesses[i] = value
	p.totalFitness += value - oldValue
}
