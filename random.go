package moea

import "math/rand"

const (
	MaxUint32     = ^uint32(0)
	HalfMaxUint32 = MaxUint32 >> 1
)

var (
	x uint32 = 123456789
	y uint32 = 362436069
	z uint32 = 521288629
	w uint32 = 88675123
	t uint32
)

func XorshiftSeed(seed uint32) {
	w = seed
}

func xorshift() uint32 {
	t = x ^ (x << 11)
	x = y
	y = z
	z = w
	w = w ^ (w >> 19) ^ (t ^ (t >> 8))
	return w
}

func flipXorshift(probability uint32) bool {
	return xorshift() < probability
}

func fairFlipXorshift() bool {
	return flipXorshift(HalfMaxUint32)
}

func flip(probability float64) bool {
	return rand.Float64() < probability
}
