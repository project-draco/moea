package moea

import (
	"math/big"
	"reflect"
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
	// TODO: test with strings with representations having more than a word
}

func TestCopy(t *testing.T) {
	i1 := newFromString([]string{"1111"})
	i2 := newFromString([]string{"0000"})
	c := i1.Copy(i2, 2, 4)
	assertEqual(t, "1100", c.(*binaryIndividual).String())
	c = i1.Copy(nil, 4, 4)
	assertEqual(t, "1111", c.(*binaryIndividual).String())
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
