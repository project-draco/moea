package integer

import "project-draco.io/moea"

type integerPopulation []moea.Individual

type integerIndividual struct {
	arr    []int
	bounds []Bound
	rng    moea.RNG
}

type Bound struct{ Min, Max int }

func (p integerPopulation) Len() int { return len(p) }

func (p integerPopulation) Individual(i int) moea.Individual { return p[i] }

func (p integerPopulation) Clone() moea.Population {
	result := make(integerPopulation, p.Len())
	copy(result, p)
	return result
}

func NewRandomIntegerPopulation(size int, length int, bounds []Bound, rng moea.RNG) moea.Population {
	result := make(integerPopulation, size)
	for i := 0; i < size; i++ {
		result[i] = newIntegerIndividual(length, bounds, rng)
	}
	return result
}

func newIntegerIndividual(length int, bounds []Bound, rng moea.RNG) moea.Individual {
	result := integerIndividual{make([]int, length), bounds, rng}
	for i := 0; i < length; i++ {
		result.arr[i] = randomInt(bounds[i], rng)
	}
	return result
}

func (ii integerIndividual) Len() int {
	return len(ii.arr)
}

func (ii integerIndividual) Value(i int) interface{} {
	return ii.arr[i]
}

func (ii integerIndividual) Copy(individual moea.Individual, start, end int) {
	other := individual.(integerIndividual)
	copy(other.arr[start:end], ii.arr[start:end])
}

func (ii integerIndividual) Mutate(mutations []int) {
	for _, v := range mutations {
		ii.arr[v] = randomInt(ii.bounds[v], ii.rng)
	}
}

func (ii integerIndividual) Clone() moea.Individual {
	result := integerIndividual{make([]int, len(ii.arr)), ii.bounds, ii.rng}
	copy(result.arr, ii.arr)
	return result
}

func randomInt(bound Bound, rng moea.RNG) int {
	return bound.Min + int(float64(bound.Max-bound.Min)*rng.Float64())
}
