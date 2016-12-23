package moea

import (
	"fmt"
	"math/big"
	"unsafe"
)

type binaryIndividual struct {
	representation []big.Word
	lengths        []int
	starts         []int
	totalLen       int
	variables      []big.Int // TODO: use pointers and lazy initialization
}

const wordBitsize = int(8 * unsafe.Sizeof(big.Word(0)))

func NewRandomBinaryPopulation(size int, lengths []int) Population {
	totalLen := 0
	starts := make([]int, len(lengths))
	for i, l := range lengths {
		starts[i] = totalLen
		totalLen += l
	}
	individualSize := totalLen / wordBitsize
	if totalLen%wordBitsize > 0 {
		individualSize++
	}
	arr := make([]big.Word, individualSize*size)
	bi := make([]binaryIndividual, size)
	result := newPopulation(size)
	vars := make([]big.Int, len(lengths)*size)
	for i := 0; i < size; i++ {
		bi[i].representation = newBinaryEncoding(arr[i*individualSize : (i+1)*individualSize])
		bi[i].lengths = lengths
		bi[i].starts = starts
		bi[i].totalLen = totalLen
		bi[i].variables = vars[i*len(lengths) : (i+1)*len(lengths)]
		result.setIndividual(&bi[i], i)
	}
	return result
}

func newBinaryEncoding(representation []big.Word) []big.Word {
	for i := 0; i < len(representation); i++ {
		for j := 0; j < wordBitsize; j++ {
			if flip(0.5) {
				representation[i] |= (1 << uint(j))
			} else {
				representation[i] &= ^(1 << uint(j))
			}
		}
	}
	return representation
}

func (r *binaryIndividual) Len() int {
	return r.totalLen
}

func (r *binaryIndividual) Value(i int) interface{} {
	start := r.starts[i] / wordBitsize
	end := (r.starts[i]+r.lengths[i])/wordBitsize + 1
	r.variables[i].SetBits(r.representation[start:end])
	r.variables[i].Rsh(&r.variables[i], uint(r.starts[i]%wordBitsize))
	return r.variables[i].Bits()
}

func (r *binaryIndividual) Copy(individual Individual, start, end int) Individual {
	result := &binaryIndividual{
		make([]big.Word, len(r.representation)),
		r.lengths,
		r.starts,
		r.totalLen,
		make([]big.Int, len(r.variables))}
	pos := 0
	for i := 0; i < len(r.representation); i++ {
		if start >= pos+wordBitsize || end <= pos {
			result.representation[i] = r.representation[i]
		} else {
			if start <= pos && end >= pos+wordBitsize && individual != nil {
				result.representation[i] = individual.(*binaryIndividual).representation[i]
			} else {
				ii := start - pos
				if ii < 0 {
					ii = 0
				}
				jj := end - pos
				if jj > wordBitsize {
					jj = wordBitsize
				}
				result.representation[i] = r.representation[i]
				setbits(result.representation[i],
					individual.(*binaryIndividual).representation[i], uint(ii), uint(jj-ii))
			}
		}
		pos += wordBitsize
	}
	return result
}

func (r *binaryIndividual) Mutate(mutations []bool) {
	for i := 0; i < len(r.representation); i++ {
		for j := 0; j < wordBitsize; j++ {
			pos := i*wordBitsize + j
			if len(mutations) > pos && mutations[pos] {
				r.representation[i] ^= 1 << uint(j)
			}
		}
	}
}

func (r *binaryIndividual) String() string {
	rmd := r.Len() % wordBitsize
	if rmd != 0 {
		mask := big.Word((1 << uint(rmd)) - 1)
		r.representation[len(r.representation)-1] &= mask
	}
	return fmt.Sprintf("%b", r.representation)
}

func setbits(destination, source big.Word, at, numbits uint) big.Word {
	mask := big.Word(((^uint(0)) >> (32 - numbits)) << at)
	return (destination &^ mask) | ((source << at) & mask)
}
