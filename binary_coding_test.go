package moea

import (
	"reflect"
	"testing"
)

func TestCopy(t *testing.T) {
	i1 := binaryIndividual([]bool{true, true, true, true})
	i2 := binaryIndividual([]bool{false, false, false, false})
	c := i1.Copy(i2, 2, 4)
	expected := binaryIndividual{true, true, false, false}
	if !reflect.DeepEqual(c, expected) {
		t.Errorf("expected %v but was %v", expected, c)
	}
	c = i1.Copy(nil, 4, 4)
	expected = binaryIndividual{true, true, true, true}
	if !reflect.DeepEqual(c, expected) {
		t.Errorf("expected %v but was %v", expected, c)
	}
}
