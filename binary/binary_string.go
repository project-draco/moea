package binary

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type BinaryString interface {
	Len() int
	Test(i int) bool
	Set(i int)
	Clear(i int)
	Flip(i int)
	Int() *big.Int
	String() string
	Slice(i, j int, dest BinaryString)
	Copy(bs BinaryString, start, end int)
	Iterator(w, j *int) BinaryStringIterator
	SetString(string)
}

type BinaryStringIterator interface {
	Test(w, j int) bool
	Set(w, j int)
	Clear(w, j int)
	Flip(w, j int)
	Next(w, j *int)
}

type bs struct {
	w       []big.Word
	bigbits []big.Word
	bigint  *big.Int
	l       int
	bsi     *bsi
}

type bsi struct{ *bs }

type pool struct {
	w              []big.Word
	bigints        []big.Int
	bigbits        []big.Word
	starts         []int
	bsis           []bsi
	lengths        []int
	count          int
	wordCountTotal int
}

func newBinString(l int, w []big.Word, bigint *big.Int, bigbits []big.Word, bsi_ *bsi) *bs {
	result := &bs{}
	result.init(l, w, bigint, bigbits, bsi_)
	return result
}

func (b *bs) init(l int, w []big.Word, bigint *big.Int, bigbits []big.Word, bsi_ *bsi) {
	words := howManyWords(l)
	if w == nil {
		w = make([]big.Word, words)
	}
	if bigbits == nil {
		bigbits = make([]big.Word, words)
		bigint = new(big.Int)
		bigint.SetBits(bigbits)
	}
	if bsi_ == nil {
		bsi_ = &bsi{}
	}
	b.w = w
	b.l = l
	b.bigbits = bigbits
	b.bigint = bigint
	b.bsi = bsi_
	b.bsi.bs = b
}

func newPool(lengths []int, count int) *pool {
	result := &pool{}
	result.starts = make([]int, len(lengths)+1)
	wordCountTotal := 0
	for i, l := range lengths {
		w := howManyWords(l)
		result.starts[i+1] = result.starts[i] + w
		wordCountTotal += w
	}
	result.w = make([]big.Word, wordCountTotal*count)
	result.bigbits = make([]big.Word, wordCountTotal*count)
	result.bigints = make([]big.Int, len(lengths)*count)
	result.bsis = make([]bsi, len(lengths)*count)
	result.lengths = lengths
	result.count = count
	result.wordCountTotal = wordCountTotal
	return result
}

func (p *pool) get(i, j int) ([]big.Word, *big.Int, []big.Word, *bsi) {
	f, t := p.starts[i]+j*p.wordCountTotal, p.starts[i+1]+j*p.wordCountTotal
	index := i + j*len(p.lengths)
	return p.w[f:t], &p.bigints[index], p.bigbits[f:t], &p.bsis[index]
}

func (p *pool) clone() *pool {
	result := newPool(p.lengths, p.count)
	copy(result.w, p.w)
	copy(result.bigbits, p.bigbits)
	copy(result.bigints, p.bigints)
	copy(result.bsis, p.bsis)
	copy(result.starts, p.starts)
	copy(result.lengths, p.lengths)
	result.count = p.count
	result.wordCountTotal = p.wordCountTotal
	return result
}

func (b *bs) Len() int {
	return b.l
}

func (b *bs) Test(i int) bool {
	var w, j int
	b.bsi.setPosition(i, &w, &j)
	return b.bsi.Test(w, j)
}

func (b *bs) Set(i int) {
	var w, j int
	b.bsi.setPosition(i, &w, &j)
	b.bsi.Set(w, j)
}

func (b *bs) Clear(i int) {
	var w, j int
	b.bsi.setPosition(i, &w, &j)
	b.bsi.Clear(w, j)
}

func (b *bs) Flip(i int) {
	var w, j int
	b.bsi.setPosition(i, &w, &j)
	b.bsi.Flip(w, j)
}

func (b *bs) Int() *big.Int {
	for i := 0; i < len(b.bigbits); i++ {
		b.bigbits[i] = 0
	}
	for i := 0; i < len(b.w)-1; i++ {
		b.bigbits[len(b.w)-2-i] = b.w[i]
	}
	if len(b.w) > 1 {
		b.bigint.SetBits(b.bigbits[0 : len(b.w)-1])
		b.bigint = b.bigint.Lsh(b.bigint, uint(b.l%wordBitsize))
		b.bigbits = b.bigint.Bits()
	}
	if len(b.w) > 0 && len(b.bigbits) > 0 {
		b.bigbits[0] += b.w[len(b.w)-1]
	}
	b.bigint.SetBits(b.bigbits)
	return b.bigint
}

func (b *bs) String() string {
	rmd := b.l % wordBitsize
	if rmd != 0 {
		mask := big.Word((1 << uint(rmd)) - 1)
		b.w[len(b.w)-1] &= mask
	}
	result := ""
	for i := 0; i < len(b.w)-1; i++ {
		s := fmt.Sprintf("%b", b.w[i])
		if len(s) < wordBitsize {
			s = strings.Repeat("0", wordBitsize-len(s)) + s
		}
		result += s
	}
	s := fmt.Sprintf("%b", b.w[len(b.w)-1])
	if len(result)+len(s) < b.l {
		s = strings.Repeat("0", b.l-len(result)-len(s)) + s
	}
	result += s
	return result
}

func (s *bs) Slice(i, j int, dest BinaryString) {
	firstWord := i / wordBitsize
	irmd := uint(i % wordBitsize)
	lastWord := firstWord + (j-i-1)/wordBitsize + 1
	destWords := dest.(*bs).w
	for k := firstWord; k < lastWord; k++ {
		destWords[k-firstWord] = (s.w[k] & (^big.Word(0) >> irmd)) << irmd
		// if the source word was shifted then copy the remaining bits from next word
		if irmd > 0 && k < len(s.w)-1 {
			nextWord := s.w[k+1]
			// if the next word is the last then shift it to fill a whole word
			if k+2 == len(s.w) && s.Len()%wordBitsize > 0 {
				nextWord <<= uint(wordBitsize - s.Len()%wordBitsize)
			}
			destWords[k-firstWord] += nextWord >> (uint(wordBitsize) - irmd)
		}
	}
	size := wordBitsize
	if lastWord == len(s.w) && s.Len()%wordBitsize > 0 {
		size = s.Len() % wordBitsize
	}
	// fill with zeroes the left of last word
	destWords[lastWord-firstWord-1] >>= uint(size - (j-i)%wordBitsize)
	if lastWord == len(s.w) && s.Len()%wordBitsize > 0 {
		destWords[lastWord-firstWord-1] &= ^big.Word(0) >> uint(wordBitsize-(j-i)%wordBitsize)
	}
	// discard unused words at the end
	wordCount := howManyWords(j - i)
	if len(destWords) > wordCount {
		destWords = destWords[0 : len(destWords)-1]
	}
	dest.(*bs).w = destWords
}

func (b *bs) Copy(other BinaryString, start, end int) {
	pos := 0
	o := other.(*bs)
	for i := 0; i < len(b.w); i++ {
		if start >= pos+wordBitsize || end <= pos {
			pos += wordBitsize
			continue
		}
		if start <= pos && end >= pos+wordBitsize {
			b.w[i] = o.w[i]
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
			if i < len(b.w)-1 {
				ii = wordBitsize - ii - ll
			} else if b.l%wordBitsize > 0 {
				ii = b.l%wordBitsize - ii - ll
			} else {
				ii = wordBitsize - ii - ll
			}
			b.w[i] = setbits(b.w[i], o.w[i], uint(ii), uint(ll))
		}
		pos += wordBitsize
	}
}

func (b *bs) Iterator(w, j *int) BinaryStringIterator {
	*w = 0
	*j = wordBitsize
	if len(b.w) == 1 {
		*j = b.l
	}
	return b.bsi
}

func (b *bsi) Next(w, j *int) {
	if *j == 0 {
		*j = wordBitsize - 1
		*w++
		if *w == len(b.w)-1 {
			*j = b.l%wordBitsize - 1
		}
	} else {
		*j--
	}
}

func (b *bsi) Test(w, j int) bool {
	return b.w[w]&(1<<uint(j)) != 0
}

func (b *bsi) Set(w, j int) {
	b.w[w] |= (1 << uint(j))
}

func (b *bsi) Clear(w, j int) {
	b.w[w] &= ^(1 << uint(j))
}

func (b *bsi) Flip(w, j int) {
	b.w[w] ^= 1 << uint(j)
}

func (b *bs) SetString(s string) {
	start := 0
	i := 0
	for start < len(s) {
		end := start + wordBitsize
		if end > len(s) {
			end = len(s)
		}
		n, err := strconv.ParseUint(s[start:end], 2, wordBitsize)
		if err != nil {
			panic(err.Error())
		}
		b.w[i] = big.Word(n)
		start += wordBitsize
		i++
	}
	b.l = len(s)
}

func (b *bsi) setPosition(i int, w, j *int) {
	*w = i / wordBitsize
	if *w == len(b.w)-1 {
		l := (b.l % wordBitsize)
		if l == 0 {
			l = wordBitsize
		}
		*j = l - i%l - 1
	} else {
		*j = wordBitsize - i%wordBitsize - 1
	}
}

func setbits(destination, source big.Word, at, numbits uint) big.Word {
	mask := big.Word(((^uint(0)) >> (uint(wordBitsize) - numbits)) << at)
	return (destination &^ mask) | (source & mask)
}
