package bloomfilter

import (
	"github.com/nnurry/probabilistics/v2/utilities/hasher"
	"github.com/nnurry/probabilistics/v2/utilities/register"
)

type CountingBF[T hasher.HashOutType] struct {
	cap    uint
	k      uint
	bitR   *register.BitRegister
	countR register.Register
	h      hasher.HashGenerator[T]
}

func (f *CountingBF[T]) Cap() uint        { return f.cap }
func (f *CountingBF[T]) HashAttr() string { return f.h.String() }

func (f *CountingBF[T]) Add(data []byte) *CountingBF[T] {
	hashes, _ := f.h.GenerateHash(data, 0, f.cap, f.k)
	for _, hash := range hashes {
		rIdx := uint(hash % T(f.cap))
		f.bitR.Write(rIdx, 1)
		f.countR.Increment(rIdx)
	}
	return f
}

func (f *CountingBF[T]) Remove(data []byte) *CountingBF[T] {
	hashes, _ := f.h.GenerateHash(data, 0, f.cap, f.k)
	for _, hash := range hashes {
		rIdx := uint(hash % T(f.cap))
		_, after, _ := f.countR.Decrement(rIdx)
		if after == 0 {
			f.bitR.Write(rIdx, 0)
		}
	}
	return f
}

func (f *CountingBF[T]) Contains(data []byte) bool {
	hashes, _ := f.h.GenerateHash(data, 0, f.cap, f.k)
	for _, hash := range hashes {
		rIdx := uint(hash % T(f.cap))
		v, err := f.bitR.Read(rIdx)
		if err != nil || v == 0 {
			return false
		}
	}
	return true
}
