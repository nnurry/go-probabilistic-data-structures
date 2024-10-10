/*
Package bloom provides data structures and methods for creating Bloom filters.

A Bloom filter is a representation of a set of _n_ items, where the main
requirement is to make membership queries; _i.e._, whether an item is a
member of a set.

A Bloom filter has two parameters: _m_, a maximum size (typically a reasonably large
multiple of the cardinality of the set to represent) and _k_, the number of hashing
functions on elements of the set. (The actual hashing functions are important, too,
but this is not a parameter for this implementation). A Bloom filter is backed by
a BitSet; a key is represented in the filter by setting the bits at each value of the
hashing functions (modulo _m_). Set membership is done by _testing_ whether the
bits at each value of the hashing functions (again, modulo _m_) are set. If so,
the item is in the set. If the item is actually in the set, a Bloom filter will
never fail (the true positive rate is 1.0); but it is susceptible to false
positives. The art is to choose _k_ and _m_ correctly.

In this implementation, the hashing functions used is murmurhash,
a non-cryptographic hashing function.

This implementation accepts keys for setting as testing as []byte. Thus, to
add a string item, "Love":

	uint n = 1000
	filter := bloom.New(20*n, 5) // load of 20, 5 keys
	filter.Add([]byte("Love"))

Similarly, to test if "Love" is in bloom:

	if filter.Test([]byte("Love"))

For numeric data, I recommend that you look into the binary/encoding library. But,
for example, to add a uint32 to the filter:

	i := uint32(100)
	n1 := make([]byte,4)
	binary.BigEndian.PutUint32(n1,i)
	f.Add(n1)

Finally, there is a method to estimate the false positive rate of a
Bloom filter with _m_ bits and _k_ hashing functions for a set of size _n_:

	if bloom.EstimateFalsePositiveRate(20*n, 5, n) > 0.001 ...

You can use it to validate the computed m, k parameters:

	m, k := bloom.EstimateParameters(n, fp)
	ActualfpRate := bloom.EstimateFalsePositiveRate(m, k, n)

or

	f := bloom.NewWithEstimates(n, fp)
	ActualfpRate := bloom.EstimateFalsePositiveRate(f.m, f.k, n)

You would expect ActualfpRate to be close to the desired fp in these cases.

The EstimateFalsePositiveRate function creates a temporary Bloom filter. It is
also relatively expensive and only meant for validation.
*/
package hasher

import (
	"encoding/binary"
	"math/bits"
	"unsafe"
)

const (
	c1_128     = 0x87c37b91114253d5
	c2_128     = 0x4cf5ad432745937f
	block_size = 16
)

// digest128 represents a partial evaluation of a 128 bites hash.
type digest128 struct {
	h1 uint64 // Unfinalized running hash part 1.
	h2 uint64 // Unfinalized running hash part 2.
}

// bmix will hash blocks (16 bytes)
func (d *digest128) bmix(p []byte) {
	nblocks := len(p) / block_size
	for i := 0; i < nblocks; i++ {
		b := (*[16]byte)(unsafe.Pointer(&p[i*block_size]))
		k1, k2 := binary.LittleEndian.Uint64(b[:8]), binary.LittleEndian.Uint64(b[8:])
		d.bmix_words(k1, k2)
	}
}

// bmix_words will hash two 64-bit words (16 bytes)
func (d *digest128) bmix_words(k1, k2 uint64) {
	h1, h2 := d.h1, d.h2

	k1 *= c1_128
	k1 = bits.RotateLeft64(k1, 31)
	k1 *= c2_128
	h1 ^= k1

	h1 = bits.RotateLeft64(h1, 27)
	h1 += h2
	h1 = h1*5 + 0x52dce729

	k2 *= c2_128
	k2 = bits.RotateLeft64(k2, 33)
	k2 *= c1_128
	h2 ^= k2

	h2 = bits.RotateLeft64(h2, 31)
	h2 += h1
	h2 = h2*5 + 0x38495ab5
	d.h1, d.h2 = h1, h2
}

// sum128 computers two 64-bit hash value. It is assumed that
// bmix was first called on the data to process complete blocks
// of 16 bytes. The 'tail' is a slice representing the 'tail' (leftover
// elements, fewer than 16). If pad_tail is true, we make it seem like
// there is an extra element with value 1 appended to the tail.
// The length parameter represents the full length of the data (including
// the blocks of 16 bytes, and, if pad_tail is true, an extra byte).
func (d *digest128) sum128(pad_tail bool, length uint, tail []byte) (h1, h2 uint64) {
	h1, h2 = d.h1, d.h2

	var k1, k2 uint64
	if pad_tail {
		switch (len(tail) + 1) & 15 {
		case 15:
			k2 ^= uint64(1) << 48
			break
		case 14:
			k2 ^= uint64(1) << 40
			break
		case 13:
			k2 ^= uint64(1) << 32
			break
		case 12:
			k2 ^= uint64(1) << 24
			break
		case 11:
			k2 ^= uint64(1) << 16
			break
		case 10:
			k2 ^= uint64(1) << 8
			break
		case 9:
			k2 ^= uint64(1) << 0

			k2 *= c2_128
			k2 = bits.RotateLeft64(k2, 33)
			k2 *= c1_128
			h2 ^= k2

			break

		case 8:
			k1 ^= uint64(1) << 56
			break
		case 7:
			k1 ^= uint64(1) << 48
			break
		case 6:
			k1 ^= uint64(1) << 40
			break
		case 5:
			k1 ^= uint64(1) << 32
			break
		case 4:
			k1 ^= uint64(1) << 24
			break
		case 3:
			k1 ^= uint64(1) << 16
			break
		case 2:
			k1 ^= uint64(1) << 8
			break
		case 1:
			k1 ^= uint64(1) << 0
			k1 *= c1_128
			k1 = bits.RotateLeft64(k1, 31)
			k1 *= c2_128
			h1 ^= k1
		}

	}
	switch len(tail) & 15 {
	case 15:
		k2 ^= uint64(tail[14]) << 48
		fallthrough
	case 14:
		k2 ^= uint64(tail[13]) << 40
		fallthrough
	case 13:
		k2 ^= uint64(tail[12]) << 32
		fallthrough
	case 12:
		k2 ^= uint64(tail[11]) << 24
		fallthrough
	case 11:
		k2 ^= uint64(tail[10]) << 16
		fallthrough
	case 10:
		k2 ^= uint64(tail[9]) << 8
		fallthrough
	case 9:
		k2 ^= uint64(tail[8]) << 0

		k2 *= c2_128
		k2 = bits.RotateLeft64(k2, 33)
		k2 *= c1_128
		h2 ^= k2

		fallthrough

	case 8:
		k1 ^= uint64(tail[7]) << 56
		fallthrough
	case 7:
		k1 ^= uint64(tail[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(tail[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(tail[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(tail[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(tail[0]) << 0
		k1 *= c1_128
		k1 = bits.RotateLeft64(k1, 31)
		k1 *= c2_128
		h1 ^= k1
	}

	h1 ^= uint64(length)
	h2 ^= uint64(length)

	h1 += h2
	h2 += h1

	h1 = _fmix64(h1)
	h2 = _fmix64(h2)

	h1 += h2
	h2 += h1

	return h1, h2
}

func _fmix64(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}

// sum256 will compute 4 64-bit hash values from the input.
// It is designed to never allocate memory on the heap. So it
// works without any byte buffer whatsoever.
// It is designed to be strictly equivalent to
//
//				a1 := []byte{1}
//	         hasher := murmur3.New128()
//	         hasher.Write(data) // #nosec
//	         v1, v2 := hasher.Sum128()
//	         hasher.Write(a1) // #nosec
//	         v3, v4 := hasher.Sum128()
//
// See TestHashRandom.
func (d *digest128) sum256(data []byte) (hash1, hash2, hash3, hash4 uint64) {
	// We always start from zero.
	d.h1, d.h2 = 0, 0
	// Process as many bytes as possible.
	d.bmix(data)
	// We have enough to compute the first two 64-bit numbers
	length := uint(len(data))
	tail_length := length % block_size
	tail := data[length-tail_length:]
	hash1, hash2 = d.sum128(false, length, tail)
	// Next we want to 'virtually' append 1 to the input, but,
	// we do not want to append to an actual array!!!
	if tail_length+1 == block_size {
		// We are left with no tail!!!
		word1 := binary.LittleEndian.Uint64(tail[:8])
		word2 := uint64(binary.LittleEndian.Uint32(tail[8 : 8+4]))
		word2 = word2 | (uint64(tail[12]) << 32) | (uint64(tail[13]) << 40) | (uint64(tail[14]) << 48)
		// We append 1.
		word2 = word2 | (uint64(1) << 56)
		// We process the resulting 2 words.
		d.bmix_words(word1, word2)
		tail := data[length:] // empty slice, deliberate.
		hash3, hash4 = d.sum128(false, length+1, tail)
	} else {
		// We still have a tail (fewer than 15 bytes) but we
		// need to append '1' to it.
		hash3, hash4 = d.sum128(true, length+1, tail)
	}

	return hash1, hash2, hash3, hash4
}
