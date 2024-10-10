package bloomfilter

import (
	"github.com/nnurry/probabilistics/v2/utilities/hasher"
	"github.com/nnurry/probabilistics/v2/utilities/register"
)

type ClassicBFBuilder[T hasher.HashOutType] struct {
	cap uint
	k   uint
	r   *register.BitRegister
	h   hasher.HashGenerator[T]
}

func NewClassicBFBuilder[T hasher.HashOutType]() *ClassicBFBuilder[T] {
	defaultCap, defaultK := ClassicBFEstimateParams(0.01, 10000)
	defaultHasher, _ := hasher.NewHashGenerator[T]("murmur3Hash128Default", 64, 128, "standard")
	defaultRegister, _ := register.NewRegister(defaultCap, 1)
	return &ClassicBFBuilder[T]{
		cap: defaultCap,
		k:   defaultK,
		r:   defaultRegister.(*register.BitRegister),
		h:   *defaultHasher,
	}
}

func (b *ClassicBFBuilder[T]) SetCap(cap uint) *ClassicBFBuilder[T] {
	b.cap = cap
	return b
}

func (b *ClassicBFBuilder[T]) SetHashNum(k uint) *ClassicBFBuilder[T] {
	b.k = k
	return b
}

func (b *ClassicBFBuilder[T]) SetRegister(r *register.BitRegister) *ClassicBFBuilder[T] {
	b.r = r
	return b
}

func (b *ClassicBFBuilder[T]) SetHashGenerator(hashFamily string, platformBit uint, outputBit uint, generateMethod string) *ClassicBFBuilder[T] {
	hashGenerator, err := hasher.NewHashGenerator[T](hashFamily, platformBit, outputBit, generateMethod)
	if err != nil {
		return b
	}
	b.h = *hashGenerator
	return b
}

func (b *ClassicBFBuilder[T]) Build() *ClassicBF[T] {
	bf := &ClassicBF[T]{
		cap: b.cap,
		k:   b.k,
		r:   b.r,
		h:   b.h,
	}
	return bf
}
