package binary

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"../"
)

func TestNewFromString(t *testing.T) {
	bi := newFromString([]string{"11", "0000", "111"}, nil)
	assertEqual(t, "110000111", bi.String())
	assertEqual(t, 9, bi.totalLen)
	assertEqual(t, []int{2, 4, 3}, bi.lengths)
	assertEqual(t, []big.Word{0x0187}, bi.representation.(*bs).w)
	assertEqual(t, []int{0, 2, 6}, bi.starts)
	assertEqual(t, big.NewInt(3), bi.Value(0).(BinaryString).Int())
	assertEqual(t, big.NewInt(0).Bytes(), bi.Value(1).(BinaryString).Int().Bytes())
	assertEqual(t, big.NewInt(7), bi.Value(2).(BinaryString).Int())
	bi = newFromString([]string{"1111", strings.Repeat("0", wordBitsize), "1111"}, nil)
	assertEqual(t, "1111"+strings.Repeat("0", wordBitsize)+"1111", bi.String())
	assertEqual(t, wordBitsize+8, bi.totalLen)
	assertEqual(t, []int{4, wordBitsize, 4}, bi.lengths)
	assertEqual(t, []big.Word{0xf << uint(wordBitsize-4), 0xf}, bi.representation.(*bs).w)
	assertEqual(t, []int{0, 4, wordBitsize + 4}, bi.starts)
	assertEqual(t, big.NewInt(0xf), bi.Value(0).(BinaryString).Int())
	assertEqual(t, big.NewInt(0).Bytes(), bi.Value(1).(BinaryString).Int().Bytes())
	assertEqual(t, big.NewInt(0xf), bi.Value(2).(BinaryString).Int())
	bi = newFromString([]string{"1111", strings.Repeat("0", wordBitsize+1), "111"}, nil)
	assertEqual(t, "1111"+strings.Repeat("0", wordBitsize+1)+"111", bi.String())
	assertEqual(t, wordBitsize+8, bi.totalLen)
	assertEqual(t, []int{4, wordBitsize + 1, 3}, bi.lengths)
	assertEqual(t, []big.Word{0xf << uint(wordBitsize-4), 0x7}, bi.representation.(*bs).w)
	assertEqual(t, []int{0, 4, wordBitsize + 5}, bi.starts)
	assertEqual(t, big.NewInt(0xf), bi.Value(0).(BinaryString).Int())
	assertEqual(t, big.NewInt(0).Bytes(), bi.Value(1).(BinaryString).Int().Bytes())
	assertEqual(t, big.NewInt(0x7), bi.Value(2).(BinaryString).Int())
	bi = newFromString([]string{"1110", "1" + strings.Repeat("0", wordBitsize) + "1", "011"}, nil)
	assertEqual(t, "11101"+strings.Repeat("0", wordBitsize)+"1011", bi.String())
	assertEqual(t, wordBitsize+9, bi.totalLen)
	assertEqual(t, []int{4, wordBitsize + 2, 3}, bi.lengths)
	i := big.NewInt(1)
	i = i.Lsh(i, uint(wordBitsize+1))
	i = i.Add(i, big.NewInt(1))
	assertEqual(t, []big.Word{0x1d << uint(wordBitsize-5), 0xb}, bi.representation.(*bs).w)
	assertEqual(t, []int{0, 4, wordBitsize + 6}, bi.starts)
	assertEqual(t, big.NewInt(0xe), bi.Value(0).(BinaryString).Int())
	assertEqual(t, i, bi.Value(1).(BinaryString).Int())
	assertEqual(t, big.NewInt(3), bi.Value(2).(BinaryString).Int())
	bi = newFromString([]string{"1110", "1" + strings.Repeat("0", wordBitsize-2) + "111", "011"}, nil)
	assertEqual(t, "11101"+strings.Repeat("0", wordBitsize-2)+"111011", bi.String())
	assertEqual(t, wordBitsize+9, bi.totalLen)
	assertEqual(t, []int{4, wordBitsize + 2, 3}, bi.lengths)
	assertEqual(t, []big.Word{0x1d << uint(wordBitsize-5), 0x3b}, bi.representation.(*bs).w)
	assertEqual(t, []int{0, 4, wordBitsize + 6}, bi.starts)
	assertEqual(t, big.NewInt(0xe), bi.Value(0).(BinaryString).Int())
	i = big.NewInt(1)
	i = i.Lsh(i, uint(wordBitsize+1))
	i = i.Add(i, big.NewInt(7))
	assertEqual(t, i, bi.Value(1).(BinaryString).Int())
	assertEqual(t, big.NewInt(3), bi.Value(2).(BinaryString).Int())
}

func TestCopy(t *testing.T) {
	i1 := newFromString([]string{"1111"}, nil)
	i2 := newFromString([]string{"0000"}, nil)
	i1.Copy(i2, 2, 4)
	assertEqual(t, "1100", i1.String())
	i1.Copy(i2, 0, 4)
	assertEqual(t, "0000", i1.String())
	i0 := newFromString([]string{"1110", "1" + strings.Repeat("0", wordBitsize-2) + "111", "011"}, nil)
	i1 = newFromString([]string{"1110", "1" + strings.Repeat("0", wordBitsize-2) + "111", "011"}, nil)
	i2 = newFromString([]string{"1110", "0" + strings.Repeat("1", wordBitsize-2) + "011", "011"}, nil)
	i1.Copy(i2, 5, 6)
	assertEqual(t, "111011"+strings.Repeat("0", wordBitsize-3)+"111011", i1.String())
	i1.Copy(i0, 0, i1.Len())
	i1.Copy(i2, wordBitsize-1, wordBitsize+1)
	assertEqual(t, "11101"+strings.Repeat("0", wordBitsize-6)+"1100111011", i1.String())
	i1.Copy(i0, 0, i1.Len())
	i1.Copy(i2, wordBitsize, wordBitsize+1)
	assertEqual(t, "11101"+strings.Repeat("0", wordBitsize-5)+"100111011", i1.String())
	i1 = newFromString([]string{strings.Repeat("0", wordBitsize*3+8)}, nil)
	i2 = newFromString([]string{strings.Repeat("1", wordBitsize*3+8)}, nil)
	i1.Copy(i2, 0, wordBitsize*3+8)
	assertEqual(t, strings.Repeat("1", wordBitsize*3+8), i1.String())
	i1 = newFromString([]string{"01011101110001101000010010100100"}, nil)
	i2 = newFromString([]string{"01101001010010001100100101110111"}, nil)
	i3 := newFromString([]string{strings.Repeat("0", 32)}, nil)
	i3.Copy(i1, 0, 7)
	i3.Copy(i2, 7, 32)
	assertEqual(t, "01011101010010001100100101110111", i3.String())
	i1 = newFromString([]string{"0111011010010000100001110011101100101001110110011001010101011101"}, nil)
	i2 = newFromString([]string{"1010111010000110010111011001100010101100100110111011000100100100"}, nil)
	i3 = newFromString([]string{strings.Repeat("0", 64)}, nil)
	i3.Copy(i1, 0, 9)
	i3.Copy(i2, 9, 64)
	assertEqual(t, "0111011010000110010111011001100010101100100110111011000100100100", i3.String())
}

func TestMutate(t *testing.T) {
	i := newFromString([]string{"0000"}, nil)
	i.Mutate([]int{2})
	assertEqual(t, "0010", i.String())
	i = newFromString([]string{strings.Repeat("0", wordBitsize*3+8)}, nil)
	m := []int{}
	for i := 0; i < wordBitsize*3+8; i++ {
		m = append(m, i)
	}
	i.Mutate(m)
	assertEqual(t, strings.Repeat("1", wordBitsize*3+8), i.String())
}

func TestClone(t *testing.T) {
	p := NewRandomBinaryPopulation(1, []int{wordBitsize*3 + 8}, nil, moea.NewXorshift())
	c := p.Clone()
	assertEqual(t, p.Individual(0).(fmt.Stringer).String(), c.Individual(0).(fmt.Stringer).String())
	m := []int{0}
	c.Individual(0).Mutate(m)
	assertNotEqual(t, p.Individual(0).(fmt.Stringer).String(), c.Individual(0).(fmt.Stringer).String())
}

func TestAsBigInt(t *testing.T) {
	assertEqual(t, "1", fmt.Sprintf("%b", newFromString([]string{"1"}, nil).representation.Int()))
	assertEqual(t, "1"+strings.Repeat("0", wordBitsize),
		fmt.Sprintf("%b",
			newFromString([]string{"1" + strings.Repeat("0", wordBitsize)}, nil).representation.Int()))
}

func TestNewFromBigInt(t *testing.T) {
	i := big.NewInt(1)
	assertEqual(t, "1", newFromBigInts([]*big.Int{i}).String())
	assertEqual(t, "1"+strings.Repeat("0", wordBitsize),
		newFromBigInts([]*big.Int{i.Lsh(i, uint(wordBitsize))}).String())
}

func TestBinaryString(t *testing.T) {
	s := strings.Repeat("1", wordBitsize*3+8)
	bi := newFromString([]string{s}, nil)
	count := 0
	var w, j int
	it := bi.Value(0).(BinaryString).Iterator(&w, &j)
	for i := 0; i < bi.Len(); i++ {
		it.Next(&w, &j)
		assertEqual(t, true, it.Test(w, j))
		count++
	}
	assertEqual(t, wordBitsize*3+8, count)
	it = bi.Value(0).(BinaryString).Iterator(&w, &j)
	for i := 0; i < bi.Len(); i++ {
		it.Next(&w, &j)
		it.Clear(w, j)
	}
	assertEqual(t, strings.Repeat("0", wordBitsize*3+8), bi.Value(0).(BinaryString).String())
	it = bi.Value(0).(BinaryString).Iterator(&w, &j)
	for i := 0; i < bi.Len(); i++ {
		it.Next(&w, &j)
		it.Set(w, j)
	}
	assertEqual(t, strings.Repeat("1", wordBitsize*3+8), bi.Value(0).(BinaryString).String())
	bs := bi.Value(0).(BinaryString)
	for i := 0; i < wordBitsize*3+8; i++ {
		bs.Clear(i)
	}
	assertEqual(t, strings.Repeat("0", wordBitsize*3+8), bi.Value(0).(BinaryString).String())
}

func TestLimit(t *testing.T) {
	bi := newFromString([]string{"11"}, []Bound{{"0", "10"}})
	assertEqual(t, "10", bi.Value(0).(BinaryString).String())
	bi = newFromString([]string{"11", "00"}, []Bound{{"1", "10"}, {"1", "10"}})
	assertEqual(t, "10", bi.Value(0).(BinaryString).String())
	assertEqual(t, "01", bi.Value(1).(BinaryString).String())
	bi = newFromString([]string{"10"}, []Bound{{"0", "10"}})
	assertEqual(t, "01", bi.Value(0).(BinaryString).String())
	bi = newFromString([]string{strings.Repeat("1", wordBitsize+2)},
		[]Bound{{"0", strings.Repeat("1", wordBitsize+1)}})
	s := bi.Value(0).(BinaryString).String()
	assertEqual(t, strings.Repeat("1", wordBitsize+1), s[1:])
	assertEqual(t, "0", s[0:1])
	bi = newFromString([]string{strings.Repeat("1", 200)},
		[]Bound{{"0", strings.Repeat("1", 100)}})
	s = bi.Value(0).(BinaryString).String()
	assertEqual(t, strings.Repeat("1", 100), s[100:])
	assertEqual(t, strings.Repeat("0", 100), s[0:100])
}

func TestValueOfVariblesMultipleOfWordSize(t *testing.T) {
	ss := make([]string, 32)
	for i := 0; i < 32; i++ {
		ss[i] = strings.Repeat("1", 4)
	}
	bi := newFromString(ss, nil)
	for i := 0; i < 32; i++ {
		assertEqual(t, big.NewInt(15), bi.Value(i).(BinaryString).Int())
	}
}

func TestValueOverlapingWords(t *testing.T) {
	ss := make([]string, 14)
	for i := 0; i < 14; i++ {
		ss[i] = strings.Repeat("1", 5)
	}
	bi := newFromString(ss, nil)
	for i := 0; i < 14; i++ {
		assertEqual(t, big.NewInt(31), bi.Value(i).(BinaryString).Int())
	}
}

func assertEqual(t *testing.T, expected, value interface{}) {
	if !reflect.DeepEqual(expected, value) {
		reportError(t, "", expected, value)
	}
}

func assertNotEqual(t *testing.T, expected, value interface{}) {
	if reflect.DeepEqual(expected, value) {
		reportError(t, "not ", expected, value)
	}
}

func reportError(t *testing.T, mod string, expected, value interface{}) {
	s1 := fmt.Sprintf("%v", expected)
	s2 := fmt.Sprintf("%v", value)
	if len(s1) > 50 || len(s2) > 50 {
		t.Errorf("%sexpected\n%v\nbut was\n%v", mod, s1, s2)
	} else {
		t.Errorf("%sexpected %v but was %v", mod, s1, s2)
	}
	// debug.PrintStack()
}
