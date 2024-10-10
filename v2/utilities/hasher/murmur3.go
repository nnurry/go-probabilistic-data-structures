// heavily inspired by Appleby's MurmurHash3 C implementation
// https://github.com/aappleby/smhasher/blob/master/src/MurmurHash3.cpp
package hasher

import (
	"encoding/binary"
	"math/bits"

	"github.com/spaolacci/murmur3"
)

// constants using for block mix
const (
	blockMixMajorConst64_128_1 uint64 = 0x87c37b91114253d5
	blockMixMajorConst64_128_2 uint64 = 0x4cf5ad432745937f

	blockMixMinorConst64_128_1 uint64 = 0x52dce729
	blockMixMinorConst64_128_2 uint64 = 0x38495ab5
)

// shared constants
const (
	blockSize uint64 = 16 // 1 block consists of 16 bits
)

func fmix64(h uint64) uint64 {
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33

	return h
}

// hash into 128-bit representation of data
// (64 MSBs in 1st 64-bit hash and 64 LSBs in 2nd 64-bit hash)
func murmur3Hash128Default(data []byte, seed uint64) ([]uint64, error) {
	dataLength := uint64(len(data))
	// initialize 2 64-bit hash repr of final 128-bit output
	h1 := seed
	h2 := seed

	/*
		Perform block mix:
		- AFAIK, for each block, 8 bits are taken for mixing with 1st 64-bit hash, 2nd one will get other 8 bits:
			+ Multiplication factors are emperically chosen for better hash diffusion
			+ Rotations are made to ensure all bits contribute to hash randomness
			+ Rotation factors are not divisors of 64/32 to makes the output less predictable, maybe? (better quality and less collision)
			+ XOR operations further reduce hash predictability and make sure 2 hash halves influence each other
		- Steps:
			+ Loop over the data and extract 2 adjacent 8-bit blocks, each for a half of final hash
			+ Multiply block with 1st crazy constant
			+ Left-rotate block: shift bits to left and pad the overflow bits to the right
			+ Multiply block with 2nd crazy constant
			+ XOR on block and its respective hash half
			+ Left-rotate hash half and add with other hash half
			+ Multiply hash half with 5 and add some crazy constants
			(the same apply for other block and its hash half with different parameters)
	*/

	numBlocks := dataLength / blockSize
	for i := uint64(0); i < numBlocks; i++ {
		block := data[i*blockSize : (i+1)*blockSize]

		k1 := binary.LittleEndian.Uint64(block[:8])
		k2 := binary.LittleEndian.Uint64(block[8:])

		k1 *= blockMixMajorConst64_128_1
		k1 = bits.RotateLeft64(k1, 31)
		k1 *= blockMixMajorConst64_128_2
		h1 ^= k1

		h1 = bits.RotateLeft64(h1, 27)
		h1 += h2
		h1 = h1*5 + blockMixMinorConst64_128_1

		k2 *= blockMixMajorConst64_128_2
		k2 = bits.RotateLeft64(k2, 33)
		k2 *= blockMixMajorConst64_128_1
		h2 ^= k2

		h2 = bits.RotateLeft64(h2, 31)
		h2 += h1
		h2 = h2*5 + blockMixMinorConst64_128_2
	}

	// process leftover part of the hash, further mix for avalanche effect
	tailBlock := data[numBlocks*blockSize:]
	tailLen := len(tailBlock)
	k1 := uint64(0)
	k2 := uint64(0)

	switch tailLen {
	case 15:
		k2 ^= uint64(tailBlock[14]) << 48
		fallthrough
	case 14:
		k2 ^= uint64(tailBlock[13]) << 40
		fallthrough
	case 13:
		k2 ^= uint64(tailBlock[12]) << 32
		fallthrough
	case 12:
		k2 ^= uint64(tailBlock[11]) << 24
		fallthrough
	case 11:
		k2 ^= uint64(tailBlock[10]) << 16
		fallthrough
	case 10:
		k2 ^= uint64(tailBlock[9]) << 8
		fallthrough
	case 9:
		// repeat some steps in block mix for tail mix
		k2 ^= uint64(tailBlock[8])
		k2 *= blockMixMajorConst64_128_2
		k2 = bits.RotateLeft64(k2, 33)
		k2 *= blockMixMajorConst64_128_1
		h2 ^= k2
		fallthrough
	case 8:
		k1 ^= uint64(tailBlock[7]) << 56
		fallthrough
	case 7:
		k1 ^= uint64(tailBlock[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(tailBlock[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(tailBlock[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(tailBlock[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(tailBlock[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(tailBlock[1]) << 8
		fallthrough
	case 1:
		// repeat some steps in block mix for tail mix
		k1 ^= uint64(tailBlock[0])
		k1 *= blockMixMajorConst64_128_1
		k1 = bits.RotateLeft64(k1, 31)
		k1 *= blockMixMajorConst64_128_2
		h1 ^= k1
	}

	// hash finalization
	h1 ^= dataLength
	h2 ^= dataLength

	h1 += h2
	h2 += h1

	h1 = fmix64(h1)
	h2 = fmix64(h2)

	h1 += h2
	h2 += h1

	return []uint64{h1, h2}, nil
}

func murmur3Hash128Spaolacci(data []byte, seed uint64) ([]uint64, error) {
	hf := murmur3.New128WithSeed(uint32(seed))
	_, err := hf.Write(data)
	if err != nil {
		return nil, err
	}
	h1, h2 := hf.Sum128()
	return []uint64{h1, h2}, nil
}

func murmur3Hash64Spaolacci(data []byte, seed uint64) ([]uint64, error) {
	hf := murmur3.New64WithSeed(uint32(seed))
	_, err := hf.Write(data)
	if err != nil {
		return nil, err
	}
	return []uint64{hf.Sum64()}, nil
}

func murmur3Hash256Bnb(data []byte, seed uint64) ([]uint64, error) {
	var hf digest128
	h1, h2, h3, h4 := hf.sum256(data)
	return []uint64{h1, h2, h3, h4}, nil
}
