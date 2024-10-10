package hasher

import (
	oneOfOneXxHash "github.com/OneOfOne/xxhash"
	cespareXxHash "github.com/cespare/xxhash"
)

func xxHash64Cespare(data []byte, seed uint64) ([]uint64, error) {
	hf := cespareXxHash.New()
	_, err := hf.Write(data)
	if err != nil {
		return nil, err
	}
	return []uint64{hf.Sum64()}, nil
}

func xxHash64OneOfOne(data []byte, seed uint64) ([]uint64, error) {
	hf := oneOfOneXxHash.NewS64(seed)
	_, err := hf.Write(data)
	if err != nil {
		return nil, err
	}
	return []uint64{hf.Sum64()}, nil
}
