package moea

import (
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
}

func TestCopy(t *testing.T) {
	i1 := newFromString([]string{"1111"})
	i2 := newFromString([]string{"0000"})
	c := i1.Copy(i2, 2, 4)
	assertEqual(t, "1100", c.(*binaryIndividual).String())
	c = i1.Copy(nil, 4, 4)
	assertEqual(t, "1111", c.(*binaryIndividual).String())
	// TODO: test with a binary string larger than a big.Word
}

func TestMutate(t *testing.T) {
	i := newFromString([]string{"0000"})
	i.Mutate([]bool{false, false, true, false})
	assertEqual(t, "0010", i.String())
}

func assertEqual(t *testing.T, expected, value interface{}) {
	if !reflect.DeepEqual(expected, value) {
		t.Errorf("expected %v but was %v", expected, value)
	}
}
