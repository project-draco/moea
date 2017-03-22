package moea

import (
	"fmt"
	"math/big"
	"unsafe"
)

type binaryPopulation struct {
	individuals []Individual
	bi          []binaryIndividual
	arr         []big.Word
	vars        []big.Word
}

type binaryIndividual struct {
	representation         BinaryString
	lengths                []int
	bounds                 []Bound
	mappings               []mapping
	starts                 []int
	totalLen               int
	variables              []*bs
	variableWordCount      []int
	variableWordCountTotal int
	variablesInitialized   bool
	rng                    RNG
}

type Bound struct {
	Min, Max string
}

type mapping struct {
	min   *big.Int
	coeff *big.Rat
}

const wordBitsize = int(8 * unsafe.Sizeof(big.Word(0)))

func NewRandomBinaryPopulation(size int, lengths []int, bounds []Bound, rng RNG) Population {
	totalLen := 0
	starts := make([]int, len(lengths))
	for i, l := range lengths {
		starts[i] = totalLen
		totalLen += l
	}
	variableWordCount, variableWordCountTotal := computeVariableWordCount(lengths)
	// varsSlices := make([][]big.Word, len(lengths)*size)
	allVariables := make([]bs, len(lengths)*size)
	pointersToAllVariables := make([]*bs, len(lengths)*size)
	for i := 0; i < len(allVariables); i++ {
		pointersToAllVariables[i] = &allVariables[i]
	}
	var mappings []mapping
	if bounds != nil {
		mappings = mappingsFromBounds(bounds, lengths)
	}
	individualSize := totalLen / wordBitsize
	if totalLen%wordBitsize > 0 {
		individualSize++
	}
	result := &binaryPopulation{
		make([]Individual, size),
		make([]binaryIndividual, size),
		make([]big.Word, individualSize*size),
		make([]big.Word, variableWordCountTotal*size)}
	for i := 0; i < size; i++ {
		result.bi[i].representation =
			newBinString(totalLen, result.arr[i*individualSize:(i+1)*individualSize], nil, nil)
		randomize(result.bi[i].representation, rng)
		result.bi[i].lengths = lengths
		result.bi[i].bounds = bounds
		result.bi[i].mappings = mappings
		result.bi[i].starts = starts
		result.bi[i].totalLen = totalLen
		// result.bi[i].variables = varsSlices[i*len(lengths) : (i+1)*len(lengths)]
		result.bi[i].variables = pointersToAllVariables[i*len(lengths) : (i+1)*len(lengths)]
		mapVars(&result.bi[i], i*variableWordCountTotal, result.vars, variableWordCount)
		result.bi[i].variableWordCount = variableWordCount
		result.bi[i].variableWordCountTotal = variableWordCountTotal
		result.bi[i].rng = rng
		result.individuals[i] = &result.bi[i]
	}
	return result
}

func (p *binaryPopulation) Len() int { return len(p.individuals) }

func (p *binaryPopulation) Individual(i int) Individual { return p.individuals[i] }

func (p *binaryPopulation) Clone() Population {
	if p.Len() == 0 {
		return p
	}
	first := p.individuals[0].(*binaryIndividual)
	individualSize := first.representation.Len() / wordBitsize
	if first.representation.Len()%wordBitsize > 0 {
		individualSize++
	}
	result := &binaryPopulation{
		make([]Individual, p.Len()),
		make([]binaryIndividual, p.Len()),
		make([]big.Word, individualSize*p.Len()),
		make([]big.Word, first.variableWordCountTotal*p.Len())}
	// varsSlices := make([][]big.Word, len(first.lengths)*p.Len())
	allVariables := make([]bs, len(first.lengths)*p.Len())
	pointersToAllVariables := make([]*bs, len(first.lengths)*p.Len())
	for i := 0; i < len(allVariables); i++ {
		pointersToAllVariables[i] = &allVariables[i]
	}
	copy(result.bi, p.bi)
	copy(result.arr, p.arr)
	copy(result.vars, p.vars)
	for i := 0; i < p.Len(); i++ {
		result.bi[i].representation =
			newBinString(first.representation.Len(), result.arr[i*individualSize:(i+1)*individualSize], nil, nil)
		// result.bi[i].variables = varsSlices[i*len(first.lengths) : (i+1)*len(first.lengths)]
		result.bi[i].variables = pointersToAllVariables[i*len(first.lengths) : (i+1)*len(first.lengths)]
		mapVars(&result.bi[i], i*first.variableWordCountTotal, result.vars, first.variableWordCount)
		result.individuals[i] = &result.bi[i]
	}
	return result
}

func (r *binaryIndividual) Clone() Individual {
	result := NewRandomBinaryPopulation(1, r.lengths, r.bounds, r.rng).Individual(0)
	result.Copy(r, 0, result.Len())
	return result
}

func mapVars(bi *binaryIndividual, v int, vars []big.Word, variableWordCount []int) {
	for j := 0; j < len(bi.lengths); j++ {
		bi.variables[j].init(bi.lengths[j], vars[v:v+variableWordCount[j]], nil, nil)
		v += variableWordCount[j]
	}
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

func randomize(representation BinaryString, rng RNG) {
	var w, j int
	it := representation.Iterator(&w, &j)
	l := representation.Len()
	for i := 0; i < l; i++ {
		it.Next(&w, &j)
		if rng.FairFlip() {
			it.Set(w, j)
		} else {
			it.Clear(w, j)
		}
	}
}

func mappingsFromBounds(bounds []Bound, lengths []int) []mapping {
	result := make([]mapping, len(bounds))
	for i, b := range bounds {
		min, ok1 := new(big.Int).SetString(b.Min, 2)
		max, ok2 := new(big.Int).SetString(b.Max, 2)
		if !ok1 || !ok2 {
			panic("Invalid bounds")
		}
		interval := new(big.Int)
		interval.Set(max)
		interval.Sub(interval, min)
		fullscale := new(big.Int)
		fullscale.Exp(big.NewInt(2), big.NewInt(int64(lengths[i])), nil)
		fullscale.Sub(fullscale, big.NewInt(1))
		fullscaleAsRat := new(big.Rat).SetInt(fullscale)
		coeff := new(big.Rat)
		coeff.SetInt(interval)
		coeff.Quo(coeff, fullscaleAsRat)
		result[i] = mapping{min, coeff}
	}
	return result
}

func (r *binaryIndividual) Len() int {
	return r.totalLen
}

func (r *binaryIndividual) Value(idx int) interface{} {
	if r.variablesInitialized {
		return r.variables[idx]
	}
	var f *big.Rat
	for i := 0; i < len(r.variables); i++ {
		r.representation.Slice(r.starts[i], r.starts[i]+r.lengths[i], r.variables[i])
		if r.mappings != nil {
			if f == nil {
				f = new(big.Rat)
			}
			bigint := r.variables[i].Int()
			f.SetInt(bigint)
			f = f.Mul(f, r.mappings[i].coeff)
			if f.Denom().Int64() == 1 {
				bigint = f.Num()
			} else {
				ff := new(big.Float).SetInt(f.Num())
				ff = ff.Quo(ff, new(big.Float).SetInt(f.Denom()))
				ff.Int(bigint)
			}
			bigint = bigint.Add(bigint, r.mappings[i].min)
			if len(r.variables[i].w) > 1 {
				rmd := r.lengths[i] % wordBitsize
				w0 := bigint.Bits()[0]
				w0 = (w0 << uint(wordBitsize-rmd)) >> uint(wordBitsize-rmd)
				bigint = bigint.Rsh(bigint, uint(rmd))
				bigbits := bigint.Bits()
				for j := 0; j < len(r.variables[i].w)-1; j++ {
					r.variables[i].w[j] = bigbits[len(bigbits)-1-j]
				}
				r.variables[i].w[len(r.variables[i].w)-1] = w0
			} else {
				r.variables[i].w[0] = bigint.Bits()[0]
			}
		}
	}
	r.variablesInitialized = true
	return r.variables[idx]
}

func (r *binaryIndividual) Copy(individual Individual, start, end int) {
	bi := individual.(*binaryIndividual)
	r.representation.Copy(bi.representation, start, end)
	r.variablesInitialized = false
}

func (r *binaryIndividual) Mutate(mutations []int) {
	for i := 0; i < len(mutations); i++ {
		r.representation.Flip(mutations[i])
	}
	r.variablesInitialized = false
}

func (r *binaryIndividual) String() string {
	return r.representation.String()
}

func newFromBigInts(ints []*big.Int) *binaryIndividual {
	strings := make([]string, len(ints))
	for i, ii := range ints {
		strings[i] = fmt.Sprintf("%b", ii)
	}
	return newFromString(strings, nil)
}

func newFromString(s []string, bounds []Bound) *binaryIndividual {
	l := 0
	for _, each := range s {
		l += len(each)
	}
	bi := &binaryIndividual{
		lengths:   make([]int, len(s)),
		starts:    make([]int, len(s)),
		variables: make([]*bs, len(s)),
	}
	ss := ""
	for i, each := range s {
		bi.lengths[i] = len(each)
		bi.starts[i] = bi.totalLen
		bi.totalLen += len(each)
		ss += s[i]
	}
	bi.representation = newBinString(len(ss), nil, nil, nil)
	bi.representation.SetString(ss)
	bi.variableWordCount, bi.variableWordCountTotal = computeVariableWordCount(bi.lengths)
	vars := make([]big.Word, bi.variableWordCountTotal)
	v := 0
	for i := 0; i < len(bi.variables); i++ {
		bi.variables[i] = newBinString(bi.lengths[i], vars[v:v+bi.variableWordCount[i]], nil, nil)
		v += bi.variableWordCount[i]
	}
	if bounds != nil {
		bi.mappings = mappingsFromBounds(bounds, bi.lengths)
	}
	return bi
}
