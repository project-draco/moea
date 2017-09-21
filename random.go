package moea

type RNG interface {
	Flip(probability float64) bool
	FairFlip() bool
	Float64() float64
}

const (
	MaxUint32        = ^uint32(0)
	HalfMaxUint32    = MaxUint32 >> 1
	MaxUint32AsFloat = float64(MaxUint32)
)

func NewXorshift() RNG {
	return NewXorshiftWithSeed(88675123)
}

func NewXorshiftWithSeed(seed uint32) RNG {
	return &Xorshift{
		x: 123456789,
		y: 362436069,
		z: 521288629,
		w: seed,
	}
}

type Xorshift struct {
	x, y, z, w, t uint32
}

func (s *Xorshift) Flip(probability float64) bool {
	return s.xorshift() < uint32(probability*MaxUint32AsFloat)
}

func (s *Xorshift) FairFlip() bool {
	return s.xorshift() < HalfMaxUint32
}

func (s *Xorshift) Float64() float64 {
	return float64(s.xorshift()) / MaxUint32AsFloat
}

func (s *Xorshift) xorshift() uint32 {
	s.t = s.x ^ (s.x << 11)
	s.x = s.y
	s.y = s.z
	s.z = s.w
	s.w = s.w ^ (s.w >> 19) ^ (s.t ^ (s.t >> 8))
	return s.w
}
