package binary

import (
	"fmt"
	"math/big"
	"unsafe"

	"../"
)

type binaryPopulation struct {
	individuals        []moea.Individual
	bi                 []binaryIndividual
	vars               []big.Word
	poolRepresentation *pool
	poolVariables      *pool
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
	rng                    moea.RNG
}

type Bound struct {
	Min, Max string
}

type mapping struct {
	min   *big.Int
	coeff *big.Rat
}

const wordBitsize = int(8 * unsafe.Sizeof(big.Word(0)))

func NewRandomBinaryPopulation(size int, lengths []int, bounds []Bound, rng moea.RNG) moea.Population {
	if size%2 == 1 {
		size++
	}
	totalLen := 0
	starts := make([]int, len(lengths))
	for i, l := range lengths {
		starts[i] = totalLen
		totalLen += l
	}
	variableWordCount, variableWordCountTotal := computeVariableWordCount(lengths)
	allVariables := make([]bs, len(lengths)*size)
	pointersToAllVariables := make([]*bs, len(lengths)*size)
	for i := 0; i < len(allVariables); i++ {
		pointersToAllVariables[i] = &allVariables[i]
	}
	var mappings []mapping
	if bounds != nil {
		mappings = mappingsFromBounds(bounds, lengths)
	}
	result := &binaryPopulation{
		make([]moea.Individual, size),
		make([]binaryIndividual, size),
		make([]big.Word, variableWordCountTotal*size),
		newPool([]int{totalLen}, size),
		newPool(lengths, size),
	}
	for i := 0; i < size; i++ {
		w, bigint, bigbits, bsi := result.poolRepresentation.get(0, i)
		result.bi[i].representation = newBinString(totalLen, w, bigint, bigbits, bsi)
		randomize(result.bi[i].representation, rng)
		result.bi[i].lengths = lengths
		result.bi[i].bounds = bounds
		result.bi[i].mappings = mappings
		result.bi[i].starts = starts
		result.bi[i].totalLen = totalLen
		result.bi[i].variables = pointersToAllVariables[i*len(lengths) : (i+1)*len(lengths)]
		result.mapVars(i)
		result.bi[i].variableWordCount = variableWordCount
		result.bi[i].variableWordCountTotal = variableWordCountTotal
		result.bi[i].rng = rng
		result.individuals[i] = &result.bi[i]
	}
	return result
}

func (p *binaryPopulation) Len() int { return len(p.individuals) }

func (p *binaryPopulation) Individual(i int) moea.Individual { return p.individuals[i] }

func (p *binaryPopulation) Clone() moea.Population {
	if p.Len() == 0 {
		return p
	}
	first := p.individuals[0].(*binaryIndividual)
	result := &binaryPopulation{
		make([]moea.Individual, p.Len()),
		make([]binaryIndividual, p.Len()),
		make([]big.Word, first.variableWordCountTotal*p.Len()),
		p.poolRepresentation.clone(),
		p.poolVariables.clone(),
	}
	allVariables := make([]bs, len(first.lengths)*p.Len())
	pointersToAllVariables := make([]*bs, len(first.lengths)*p.Len())
	for i := 0; i < len(allVariables); i++ {
		pointersToAllVariables[i] = &allVariables[i]
	}
	copy(result.bi, p.bi)
	copy(result.vars, p.vars)
	for i := 0; i < p.Len(); i++ {
		w, bigint, bigbits, bsi := result.poolRepresentation.get(0, i)
		result.bi[i].representation = newBinString(first.representation.Len(), w, bigint, bigbits, bsi)
		result.bi[i].variables = pointersToAllVariables[i*len(first.lengths) : (i+1)*len(first.lengths)]
		result.mapVars(i)
		result.individuals[i] = &result.bi[i]
	}
	return result
}

func (r *binaryIndividual) Clone() moea.Individual {
	result := NewRandomBinaryPopulation(1, r.lengths, r.bounds, r.rng).Individual(0)
	result.Copy(r, 0, result.Len())
	return result
}

func (bp *binaryPopulation) mapVars(j int) {
	for i := 0; i < len(bp.bi[j].lengths); i++ {
		w, bigint, bigbits, bsi := bp.poolVariables.get(i, j)
		bp.bi[j].variables[i].init(bp.bi[j].lengths[i], w, bigint, bigbits, bsi)
	}
}

func randomize(representation BinaryString, rng moea.RNG) {
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
	for i := 0; i < len(r.variables); i++ {
		r.representation.Slice(r.starts[i], r.starts[i]+r.lengths[i], r.variables[i])
		if r.mappings != nil {
			bigint := r.variables[i].Int()
			if r.mappings[i].coeff.Num().BitLen() != 1 {
				bigint.Mul(bigint, r.mappings[i].coeff.Num())
			}
			bigint.Quo(bigint, r.mappings[i].coeff.Denom())
			bigint = bigint.Add(bigint, r.mappings[i].min)
			if len(r.variables[i].w) > 1 {
				rmd := r.lengths[i] % wordBitsize
				b0 := bigint.Bits()[0]
				b0 = (b0 << uint(wordBitsize-rmd)) >> uint(wordBitsize-rmd)
				bigint = bigint.Rsh(bigint, uint(rmd))
				bigbits := bigint.Bits()
				ll := len(r.variables[i].w) - len(bigbits)
				if rmd > 0 && ll > 0 {
					ll--
				}
				for j := ll; j < len(r.variables[i].w)-1; j++ {
					r.variables[i].w[j] = bigbits[len(bigbits)-1-j+ll]
				}
				for j := 0; j < ll; j++ {
					r.variables[i].w[j] = 0
				}
				if rmd > 0 {
					r.variables[i].w[len(r.variables[i].w)-1] = b0
				}
			} else {
				bits := bigint.Bits()
				if len(bits) == 0 {
					r.variables[i].w[0] = 0
				} else {
					r.variables[i].w[0] = bits[0]
				}
			}
		}
	}
	r.variablesInitialized = true
	return r.variables[idx]
}

func (r *binaryIndividual) Copy(individual moea.Individual, start, end int) {
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
	bi.representation = newBinString(len(ss), nil, nil, nil, nil)
	bi.representation.SetString(ss)
	bi.variableWordCount, bi.variableWordCountTotal = computeVariableWordCount(bi.lengths)
	vars := make([]big.Word, bi.variableWordCountTotal)
	v := 0
	for i := 0; i < len(bi.variables); i++ {
		bi.variables[i] = newBinString(bi.lengths[i], vars[v:v+bi.variableWordCount[i]], nil, nil, nil)
		v += bi.variableWordCount[i]
	}
	if bounds != nil {
		bi.mappings = mappingsFromBounds(bounds, bi.lengths)
	}
	return bi
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

func howManyWords(i int) int {
	result := i / wordBitsize
	if i%wordBitsize > 0 {
		result++
	}
	return result
}
