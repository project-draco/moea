package moea

type binaryPopulation struct {
	*population
	lengths []int
}

type binaryIndividual [][]bool

func NewRandomBinaryPopulation(size int, lengths []int) Population {
	result := binaryPopulation{newPopulation(size), lengths}
	for i := 0; i < size; i++ {
		result.setIndividual(newRandomBinaryCoding(lengths), i)
	}
	return result
}

func newRandomBinaryCoding(lengths []int) Individual {
	representation := make([][]bool, len(lengths))
	for i := 0; i < len(lengths); i++ {
		representation[i] = make([]bool, lengths[i])
		for j := 0; j < lengths[i]; j++ {
			representation[i][j] = flip(0.5)
		}
	}
	return binaryIndividual(representation)
}

func (r binaryIndividual) Len() int {
	result := 0
	for _, l := range r {
		result += len(l)
	}
	return result
}

func (r binaryIndividual) Value(i int) interface{} {
	return r[i]
}

func (r binaryIndividual) Copy(individual Individual, start, end int) Individual {
	result := make([][]bool, len(r))
	pos := 0
	for i := 0; i < len(r); i++ {
		if start >= pos+len(r[i]) || end <= pos {
			result[i] = r[i]
		} else {
			if start <= pos && end >= pos+len(r[i]) && individual != nil {
				result[i] = individual.Value(i).([]bool)
			} else {
				result[i] = make([]bool, len(r[i]))
				ii := start - pos
				if ii < 0 {
					ii = 0
				}
				jj := end - pos
				if jj > len(r[i]) {
					jj = len(r[i])
				}
				copy(result[i][0:ii], r[i][0:ii])
				if individual != nil {
					bi := individual.(binaryIndividual)
					copy(result[i][ii:jj], bi[i][ii:jj])
				}
				copy(result[i][jj:], r[i][jj:])
			}
		}
		pos += len(r[i])
	}
	return binaryIndividual(result)
}

func (r binaryIndividual) Mutate(mutations []bool) {
	pos := 0
	for i := 0; i < len(r); i++ {
		for j := 0; j < len(r[i]); j++ {
			if mutations[pos+j] {
				r[i][j] = !r[i][j]
			}
		}
		pos += len(r[i])
	}
}

func (r binaryIndividual) String() string {
	b := make([]byte, r.Len())
	pos := 0
	for i := 0; i < len(r); i++ {
		for j := 0; j < len(r[i]); j++ {
			if r[i][j] {
				b[pos+j] = '1'
			} else {
				b[pos+j] = '0'
			}
		}
		pos += len(r[i])
	}
	return string(b)
}
