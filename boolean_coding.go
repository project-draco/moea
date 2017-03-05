package moea

type booleanPopulation []Individual

type booleanIndividual [][]bool

func (p booleanPopulation) Len() int { return len(p) }

func (p booleanPopulation) Individual(i int) Individual { return p[i] }

func (p booleanPopulation) Clone() Population {
	result := make(booleanPopulation, p.Len())
	copy(result, p)
	return result
}

func NewRandomBooleanPopulation(size int, lengths []int) Population {
	result := make(booleanPopulation, size)
	rng := NewXorshift()
	for i := 0; i < size; i++ {
		result[i] = newBooleanIndividual(lengths, rng)
	}
	return result
}

func newBooleanIndividual(lengths []int, rng RNG) Individual {
	result := make(booleanIndividual, len(lengths))
	for i, l := range lengths {
		result[i] = make([]bool, l)
		for j := 0; j < l; j++ {
			result[i][j] = rng.FairFlip()
		}
	}
	return result
}

func (bi booleanIndividual) Len() int {
	result := 0
	for _, v := range bi {
		result += len(v)
	}
	return result
}

func (bi booleanIndividual) Value(i int) interface{} {
	return bi[i]
}

func (bi booleanIndividual) Copy(individual Individual, start, end int) {
	other := individual.(booleanIndividual)
	k := 0
	for i, v := range bi {
		for j := 0; j < len(v); j++ {
			if k >= start && k < end {
				bi[i][j] = other[i][j]
			}
			k++
		}
	}
}

func (bi booleanIndividual) Mutate(mutations []bool) {
	k := 0
	for i, v := range bi {
		for j := 0; j < len(v); j++ {
			if mutations[k] {
				bi[i][j] = !bi[i][j]
			}
			k++
		}
	}
}

func (bi booleanIndividual) Clone() Individual {
	result := make(booleanIndividual, len(bi))
	for i, v := range bi {
		result[i] = make([]bool, len(v))
		copy(result[i], v)
	}
	return result
}
