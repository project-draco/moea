package moea

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unsafe"
)

type binaryIndividual struct {
	representation         []big.Word
	lengths                []int
	starts                 []int
	totalLen               int
	variables              [][]big.Word
	variableWordCount      []int
	variableWordCountTotal int
	variablesInitialized   bool
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
	variableWordCount, variableWordCountTotal := computeVariableWordCount(lengths)
	vars := make([]big.Word, variableWordCountTotal*size)
	varsSlices := make([][]big.Word, len(lengths)*size)
	result := newPopulation(size)
	for i := 0; i < size; i++ {
		bi[i].representation = newBinaryEncoding(arr[i*individualSize : (i+1)*individualSize])
		bi[i].lengths = lengths
		bi[i].starts = starts
		bi[i].totalLen = totalLen
		bi[i].variables = varsSlices[i*len(lengths) : (i+1)*len(lengths)]
		v := i * variableWordCountTotal
		for j := 0; j < len(lengths); j++ {
			bi[i].variables[j] = vars[v : v+variableWordCount[j]]
			v += variableWordCount[j]
		}
		bi[i].variableWordCount = variableWordCount
		bi[i].variableWordCountTotal = variableWordCountTotal
		result.setIndividual(&bi[i], i)
	}
	return result
}

func computeVariableWordCount(lengths []int) ([]int, int) {
	variableWordCountTotal := 0
	variableWordCount := make([]int, len(lengths))
	for i, l := range lengths {
		variableWordCount[i] = l/wordBitsize + 1
		variableWordCountTotal += variableWordCount[i]
	}
	return variableWordCount, variableWordCountTotal
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

func (r *binaryIndividual) Value(idx int) interface{} {
	if r.variablesInitialized {
		return r.variables[idx]
	}
	for i := 0; i < len(r.variables); i++ {
		start := r.starts[i] / wordBitsize
		srmd := r.starts[i] % wordBitsize
		end := (r.starts[i]+r.lengths[i])/wordBitsize + 1
		rr := r.variables[i]
		for j := 0; j < end-start; j++ {
			rr[j] = r.representation[start+j] << uint(srmd)
			if srmd > 0 && j < end-start-1 {
				size := wordBitsize
				if start+j+2 == len(r.representation) && r.Len()%wordBitsize > 0 {
					size = r.Len() % wordBitsize
				}
				nextWord := r.representation[start+j+1] >> uint(size-srmd)
				rr[j] = setbits(rr[j], nextWord, 0, uint(srmd))
			}
		}
		size := wordBitsize
		if end == len(r.representation) && r.Len()%wordBitsize > 0 {
			size = r.Len() % wordBitsize
		}
		rr[end-start-1] >>= uint(size - r.lengths[i]%wordBitsize)
		if end == len(r.representation) && r.Len()%wordBitsize > 0 {
			rr[end-start-1] &= ^big.Word(0) >> uint(wordBitsize-r.lengths[i]%wordBitsize)
		}
		ll := r.lengths[i] / wordBitsize
		if r.lengths[i]%wordBitsize > 0 {
			ll++
		}
		if len(rr) > ll {
			rr = rr[0 : len(rr)-1]
		}
		r.variables[i] = rr
	}
	r.variablesInitialized = true
	return r.variables[idx]
}

func (r *binaryIndividual) Copy(individual Individual, start, end int) Individual {
	result := &binaryIndividual{
		make([]big.Word, len(r.representation)),
		r.lengths,
		r.starts,
		r.totalLen,
		make([][]big.Word, len(r.variables)),
		r.variableWordCount,
		r.variableWordCountTotal,
		false}
	vars := make([]big.Word, r.variableWordCountTotal)
	v := 0
	for i := 0; i < len(r.variables); i++ {
		result.variables[i] = vars[v : v+r.variableWordCount[i]]
		v += r.variableWordCount[i]
	}
	pos := 0
	var bi *binaryIndividual
	if individual != nil {
		bi = individual.(*binaryIndividual)
	}
	for i := 0; i < len(r.representation); i++ {
		if start >= pos+wordBitsize || end <= pos || individual == nil {
			result.representation[i] = r.representation[i]
		} else {
			if start <= pos && end >= pos+wordBitsize {
				result.representation[i] = bi.representation[i]
			} else {
				ii := start - pos
				if ii < 0 {
					ii = 0
				}
				jj := end - pos
				if jj > wordBitsize {
					jj = wordBitsize
				}
				ll := jj - ii
				if i < len(r.representation)-1 {
					ii = wordBitsize - ii - ll
				} else {
					ii = r.totalLen%wordBitsize - ii - ll
				}
				result.representation[i] = setbits(r.representation[i],
					bi.representation[i], uint(ii), uint(ll))
			}
		}
		pos += wordBitsize
	}
	return result
}

func (r *binaryIndividual) Mutate(mutations []bool) {
	rmd := r.totalLen % wordBitsize
	for i := 0; i < len(r.representation); i++ {
		for j := 0; j < wordBitsize; j++ {
			pos := i*wordBitsize + j
			if pos < len(mutations) && mutations[pos] {
				posj := wordBitsize - j - 1
				if i == len(r.representation)-1 {
					posj = rmd - j - 1
				}
				r.representation[i] ^= 1 << uint(posj)
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
	result := ""
	for i := 0; i < len(r.representation)-1; i++ {
		s := fmt.Sprintf("%b", r.representation[i])
		if len(s) < wordBitsize {
			s = strings.Repeat("0", wordBitsize-len(s)) + s
		}
		result += s
	}
	s := fmt.Sprintf("%b", r.representation[len(r.representation)-1])
	if len(result)+len(s) < r.totalLen {
		s = strings.Repeat("0", r.totalLen-len(result)-len(s)) + s
	}
	result += s
	return result
}

func setbits(destination, source big.Word, at, numbits uint) big.Word {
	mask := big.Word(((^uint(0)) >> (uint(wordBitsize) - numbits)) << at)
	return (destination &^ mask) | ((source << at) & mask)
}

func newFromString(s []string) *binaryIndividual {
	l := 0
	for _, each := range s {
		l += len(each)
	}
	count := l / wordBitsize
	if l%wordBitsize > 0 {
		count++
	}
	bi := &binaryIndividual{
		representation: make([]big.Word, count),
		lengths:        make([]int, len(s)),
		starts:         make([]int, len(s)),
		variables:      make([][]big.Word, len(s)),
	}
	ss := ""
	for i, each := range s {
		bi.lengths[i] = len(each)
		bi.starts[i] = bi.totalLen
		bi.totalLen += len(each)
		ss += s[i]
	}
	bi.variableWordCount, bi.variableWordCountTotal = computeVariableWordCount(bi.lengths)
	vars := make([]big.Word, bi.variableWordCountTotal)
	v := 0
	for i := 0; i < len(bi.variables); i++ {
		bi.variables[i] = vars[v : v+bi.variableWordCount[i]]
		v += bi.variableWordCount[i]
	}
	start := 0
	i := 0
	for start < len(ss) {
		end := start + wordBitsize
		if end > len(ss) {
			end = len(ss)
		}
		n, err := strconv.ParseUint(ss[start:end], 2, wordBitsize)
		if err != nil {
			panic(err.Error())
		}
		bi.representation[i] = big.Word(n)
		start += wordBitsize
		i++
	}
	return bi
}
