package bloomfilter

import (
	"github.com/nnurry/probabilistics/v2/utilities/hasher"
	"github.com/nnurry/probabilistics/v2/utilities/register"
)

type CountingBFBuilder[T hasher.HashOutType] struct {
	cap    uint
	k      uint
	bitR   *register.BitRegister
	countR register.Register
	h      hasher.HashGenerator[T]
}

func NewCountingBFBuilder[T hasher.HashOutType]() *CountingBFBuilder[T] {
	// AFAIK, classic BF and counting BF's optimal parameters are similar
	// so let's use optimization function of classic BF
	defaultCap, defaultK := ClassicBFEstimateParams(0.01, 10000)
	defaultHasher, _ := hasher.NewHashGenerator[T]("murmur3Hash128Default", 64, 128, "standard")
	defaultBitRegister, _ := register.NewRegister(defaultCap, 1)
	defaultCountRegister, _ := register.NewRegister(defaultCap, 4)
	return &CountingBFBuilder[T]{
		cap:    defaultCap,
		k:      defaultK,
		bitR:   defaultBitRegister.(*register.BitRegister),
		countR: defaultCountRegister,
		h:      *defaultHasher,
	}
}

func (b *CountingBFBuilder[T]) SetCap(cap uint) *CountingBFBuilder[T] {
	b.cap = cap
	return b
}

func (b *CountingBFBuilder[T]) SetHashNum(k uint) *CountingBFBuilder[T] {
	b.k = k
	return b
}

func (b *CountingBFBuilder[T]) SetBitRegister(r *register.BitRegister) *CountingBFBuilder[T] {
	b.bitR = r
	return b
}

func (b *CountingBFBuilder[T]) SetCountRegister(r register.Register) *CountingBFBuilder[T] {
	b.countR = r
	return b
}

func (b *CountingBFBuilder[T]) SetHashGenerator(hashFamily string, platformBit uint, outputBit uint, generateMethod string) *CountingBFBuilder[T] {
	hashGenerator, err := hasher.NewHashGenerator[T](hashFamily, platformBit, outputBit, generateMethod)
	if err != nil {
		return b
	}
	b.h = *hashGenerator
	return b
}

func (b *CountingBFBuilder[T]) Build() *CountingBF[T] {
	bf := &CountingBF[T]{
		cap:    b.cap,
		k:      b.k,
		bitR:   b.bitR,
		countR: b.countR,
		h:      b.h,
	}
	return bf
}
