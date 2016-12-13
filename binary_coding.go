package moea

type binaryIndividual []bool

func (r binaryIndividual) Len() int {
	return len(r)
}

func (r binaryIndividual) Value(i int) interface{} {
	return r[i]
}

func (r binaryIndividual) Copy(individual Individual, i, j int) Individual {
	result := make([]bool, r.Len())
	copy(result[0:i], r[0:i])
	if individual != nil {
		bi := individual.(binaryIndividual)
		copy(result[i:j], bi[i:j])
	}
	copy(result[j:], r[j:])
	return binaryIndividual(result)
}

func (r binaryIndividual) Mutate(mutations []bool) {
	for i := 0; i < r.Len(); i++ {
		if mutations[i] {
			r[i] = !r[i]
		}
	}
}

func (r binaryIndividual) String() string {
	b := make([]byte, r.Len())
	for i, v := range r {
		if v {
			b[i] = '1'
		} else {
			b[i] = '0'
		}
	}
	return string(b)
}

func NewRandomBinaryPopulation(size, encodingLen int) Population {
	result := newPopulation(size)
	for i := 0; i < size; i++ {
		result.setIndividual(newRandomBinaryCoding(encodingLen), i)
	}
	return result
}

func newRandomBinaryCoding(len int) Individual {
	representation := make([]bool, len)
	for i := 0; i < len; i++ {
		representation[i] = flip(0.5)
	}
	return binaryIndividual(representation)
}
