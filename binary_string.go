package moea

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
	w   []big.Word
	l   int
	bsi *bsi
}

type bsi struct{ *bs }

func newBinString(l int, w []big.Word) *bs {
	result := &bs{}
	result.init(l, w)
	return result
}

func (b *bs) init(l int, w []big.Word) {
	words := l / wordBitsize
	if l%wordBitsize > 0 {
		words++
	}
	if w == nil {
		w = make([]big.Word, words)
	}
	b.w = w
	b.l = l
	b.bsi = &bsi{b}
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
	i, ok := new(big.Int).SetString(b.String(), 2)
	if !ok {
		panic("invalid binary string")
	}
	return i
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

func (b *bs) Slice(i, j int, dest BinaryString) {
	start := i / wordBitsize
	srmd := i % wordBitsize
	end := j/wordBitsize + 1
	rr := dest.(*bs).w
	for j := 0; j < end-start; j++ {
		rr[j] = b.w[start+j] << uint(srmd)
		if srmd > 0 && j < end-start-1 {
			size := wordBitsize
			if start+j+2 == len(b.w) && b.Len()%wordBitsize > 0 {
				size = b.Len() % wordBitsize
			}
			nextWord := b.w[start+j+1] >> uint(size-srmd)
			rr[j] = setbits(rr[j], nextWord, 0, uint(srmd))
		}
	}
	size := wordBitsize
	if end == len(b.w) && b.Len()%wordBitsize > 0 {
		size = b.Len() % wordBitsize
	}
	rr[end-start-1] >>= uint(size - (j-i)%wordBitsize)
	if end == len(b.w) && b.Len()%wordBitsize > 0 {
		rr[end-start-1] &= ^big.Word(0) >> uint(wordBitsize-(j-i)%wordBitsize)
	}
	ll := (j - i) / wordBitsize
	if (j-i)%wordBitsize > 0 {
		ll++
	}
	if len(rr) > ll {
		rr = rr[0 : len(rr)-1]
	}
	dest.(*bs).w = rr
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
			} else {
				ii = b.l%wordBitsize - ii - ll
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
}

func (b *bsi) setPosition(i int, w, j *int) {
	*w = i / wordBitsize
	if *w == len(b.w)-1 {
		l := (b.l % wordBitsize)
		*j = l - i%l - 1
	} else {
		*j = wordBitsize - i%wordBitsize - 1
	}
}

func setbits(destination, source big.Word, at, numbits uint) big.Word {
	mask := big.Word(((^uint(0)) >> (uint(wordBitsize) - numbits)) << at)
	return (destination &^ mask) | ((source << at) & mask)
}
