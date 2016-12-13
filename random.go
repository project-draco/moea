package moea

import "math/rand"

func flip(probability float64) bool {
	return rand.Float64() < probability
}
