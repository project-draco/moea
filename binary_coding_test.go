package moea

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func TestNewFromString(t *testing.T) {
	bi := newFromString([]string{"11", "0000", "111"})
	assertEqual(t, "110000111", bi.String())
	assertEqual(t, 9, bi.totalLen)
	assertEqual(t, []int{2, 4, 3}, bi.lengths)
	assertEqual(t, []big.Word{0x0187}, bi.representation)
	assertEqual(t, []int{0, 2, 6}, bi.starts)
	assertEqual(t, []big.Word{3}, bi.Value(0))
	assertEqual(t, []big.Word{0}, bi.Value(1))
	assertEqual(t, []big.Word{7}, bi.Value(2))
	bi = newFromString([]string{"1111", strings.Repeat("0", wordBitsize), "1111"})
	assertEqual(t, "1111"+strings.Repeat("0", wordBitsize)+"1111", bi.String())
	assertEqual(t, wordBitsize+8, bi.totalLen)
	assertEqual(t, []int{4, wordBitsize, 4}, bi.lengths)
	assertEqual(t, []big.Word{0xf << uint(wordBitsize-4), 0xf}, bi.representation)
	assertEqual(t, []int{0, 4, wordBitsize + 4}, bi.starts)
	assertEqual(t, []big.Word{0xf}, bi.Value(0))
	assertEqual(t, []big.Word{0}, bi.Value(1))
	assertEqual(t, []big.Word{0xf}, bi.Value(2))
	bi = newFromString([]string{"1111", strings.Repeat("0", wordBitsize+1), "111"})
	assertEqual(t, "1111"+strings.Repeat("0", wordBitsize+1)+"111", bi.String())
	assertEqual(t, wordBitsize+8, bi.totalLen)
	assertEqual(t, []int{4, wordBitsize + 1, 3}, bi.lengths)
	assertEqual(t, []big.Word{0xf << uint(wordBitsize-4), 0x7}, bi.representation)
	assertEqual(t, []int{0, 4, wordBitsize + 5}, bi.starts)
	assertEqual(t, []big.Word{0xf}, bi.Value(0))
	assertEqual(t, []big.Word{0, 0}, bi.Value(1))
	assertEqual(t, []big.Word{0x7}, bi.Value(2))
	bi = newFromString([]string{"1110", "1" + strings.Repeat("0", wordBitsize) + "1", "011"})
	assertEqual(t, "11101"+strings.Repeat("0", wordBitsize)+"1011", bi.String())
	assertEqual(t, wordBitsize+9, bi.totalLen)
	assertEqual(t, []int{4, wordBitsize + 2, 3}, bi.lengths)
	assertEqual(t, []big.Word{0x1d << uint(wordBitsize-5), 0xb}, bi.representation)
	assertEqual(t, []int{0, 4, wordBitsize + 6}, bi.starts)
	assertEqual(t, []big.Word{0xe}, bi.Value(0))
	assertEqual(t, []big.Word{1 << uint(wordBitsize-1), 1}, bi.Value(1))
	assertEqual(t, []big.Word{3}, bi.Value(2))
	bi = newFromString([]string{"1110", "1" + strings.Repeat("0", wordBitsize-2) + "111", "011"})
	assertEqual(t, "11101"+strings.Repeat("0", wordBitsize-2)+"111011", bi.String())
	assertEqual(t, wordBitsize+9, bi.totalLen)
	assertEqual(t, []int{4, wordBitsize + 2, 3}, bi.lengths)
	assertEqual(t, []big.Word{0x1d << uint(wordBitsize-5), 0x3b}, bi.representation)
	assertEqual(t, []int{0, 4, wordBitsize + 6}, bi.starts)
	assertEqual(t, []big.Word{0xe}, bi.Value(0))
	assertEqual(t, []big.Word{(1 << uint(wordBitsize-1)) + 1, 3}, bi.Value(1))
	assertEqual(t, []big.Word{3}, bi.Value(2))
}

func TestCopy(t *testing.T) {
	i1 := newFromString([]string{"1111"})
	i2 := newFromString([]string{"0000"})
	i1.Copy(i2, 2, 4)
	assertEqual(t, "1100", i1.String())
	i1.Copy(i2, 0, 4)
	assertEqual(t, "0000", i1.String())
	i0 := newFromString([]string{"1110", "1" + strings.Repeat("0", wordBitsize-2) + "111", "011"})
	i1 = newFromString([]string{"1110", "1" + strings.Repeat("0", wordBitsize-2) + "111", "011"})
	i2 = newFromString([]string{"1110", "0" + strings.Repeat("1", wordBitsize-2) + "011", "011"})
	i1.Copy(i2, 5, 6)
	assertEqual(t, "111011"+strings.Repeat("0", wordBitsize-3)+"111011", i1.String())
	i1.Copy(i0, 0, i1.Len())
	i1.Copy(i2, wordBitsize-1, wordBitsize+1)
	assertEqual(t, "11101"+strings.Repeat("0", wordBitsize-6)+"1100111011", i1.String())
	i1.Copy(i0, 0, i1.Len())
	i1.Copy(i2, wordBitsize, wordBitsize+1)
	assertEqual(t, "11101"+strings.Repeat("0", wordBitsize-5)+"100111011", i1.String())
}

func TestMutate(t *testing.T) {
	i := newFromString([]string{"0000"})
	i.Mutate([]bool{false, false, true, false})
	assertEqual(t, "0010", i.String())
}

func TestClone(t *testing.T) {
	p := NewRandomBinaryPopulation(1, []int{1})
	c := p.Clone()
	assertEqual(t, p.Individual(0), c.Individual(0))
	assertEqual(t, p.Individual(0).Value(0), c.Individual(0).Value(0))
	c.Individual(0).Mutate([]bool{true})
	assertNotEqual(t, p.Individual(0), c.Individual(0))
	assertNotEqual(t, p.Individual(0).Value(0), c.Individual(0).Value(0))
}

func assertEqual(t *testing.T, expected, value interface{}) {
	if !reflect.DeepEqual(expected, value) {
		reportError(t, expected, value)
	}
}

func assertNotEqual(t *testing.T, expected, value interface{}) {
	if reflect.DeepEqual(expected, value) {
		reportError(t, expected, value)
	}
}

func reportError(t *testing.T, expected, value interface{}) {
	s1 := fmt.Sprintf("%v", expected)
	s2 := fmt.Sprintf("%v", value)
	if len(s1) > 50 || len(s2) > 50 {
		t.Errorf("expected\n%v\nbut was\n%v", s1, s2)
	} else {
		t.Errorf("expected %v but was %v", s1, s2)
	}
}
