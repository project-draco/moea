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
	Iterator() BinaryStringIterator
	SetString(string)
}

type BinaryStringIterator interface {
	Test() bool
	Set()
	Clear()
	Flip()
	Next() bool
}

type bs struct {
	w []big.Word
	l int
	i *bsi
}

type bsi struct {
	bs      *bs
	w, i, j int
}

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
	b.i = &bsi{}
	b.i.bs = b
}

func (b *bs) Len() int {
	return b.l
}

func (b *bs) Test(i int) bool {
	b.i.setPosition(i)
	return b.i.Test()
}

func (b *bs) Set(i int) {
	b.i.setPosition(i)
	b.i.Set()
}

func (b *bs) Clear(i int) {
	b.i.setPosition(i)
	b.i.Clear()
}

func (b *bs) Flip(i int) {
	b.i.setPosition(i)
	b.i.Flip()
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

func (b *bs) Iterator() BinaryStringIterator {
	b.i.w = 0
	b.i.i = wordBitsize
	if len(b.w) == 1 {
		b.i.i = b.l
	}
	b.i.j = 0
	return b.i
}

func (i *bsi) Next() bool {
	if i.j >= i.bs.l {
		return false
	}
	if i.i == 0 {
		i.i = wordBitsize - 1
		i.w++
		if i.w == len(i.bs.w)-1 && i.bs.l%wordBitsize > 0 {
			i.i = i.bs.l%wordBitsize - 1
		}
	} else {
		i.i--
	}
	i.j++
	return true
}

func (i *bsi) Test() bool {
	return i.bs.w[i.w]&(1<<uint(i.i)) != 0
}

func (i *bsi) Set() {
	i.bs.w[i.w] |= (1 << uint(i.i))
}

func (i *bsi) Clear() {
	i.bs.w[i.w] &= ^(1 << uint(i.i))
}

func (i *bsi) Flip() {
	i.bs.w[i.w] ^= 1 << uint(i.i)
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

func (bsi *bsi) setPosition(i int) {
	bsi.w = i / wordBitsize
	if bsi.w == len(bsi.bs.w)-1 {
		bsi.i = i % wordBitsize
	} else {
		bsi.i = wordBitsize - i%wordBitsize - 1
	}
	bsi.j = i
}

func setbits(destination, source big.Word, at, numbits uint) big.Word {
	mask := big.Word(((^uint(0)) >> (uint(wordBitsize) - numbits)) << at)
	return (destination &^ mask) | ((source << at) & mask)
}
