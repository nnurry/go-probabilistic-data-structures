package test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/nnurry/probabilistics/v2/membership/bloomfilter"
	"github.com/nnurry/probabilistics/v2/utilities/hasher"
	"github.com/nnurry/probabilistics/v2/utilities/register"
)

func testClassicBloomHelperBasic(
	fpr float64, elems uint, populationRatio float64, generateMethod string, hashFuncAttr hasher.HashAttribute) int64 {
	start := time.Now() // Start the timer

	testFp := fpr
	testN := elems
	testHashGenerateMethod := generateMethod
	realTestN := uint(float64(testN) / float64(populationRatio))

	builder := bloomfilter.NewClassicBFBuilder[uint64]()
	optM, optK := bloomfilter.ClassicBFEstimateParams(testFp, testN)

	r, _ := register.NewRegister(optM, 1)

	builder = builder.
		SetCap(optM).
		SetHashNum(optK).
		SetRegister(r.(*register.BitRegister)).
		SetHashGenerator(
			hashFuncAttr.HashFamily,
			hashFuncAttr.PlatformBit,
			hashFuncAttr.OutputBit,
			testHashGenerateMethod,
		)
	bf := builder.Build()
	log.Println("bloom:", bf)

	typeName := fmt.Sprintf("%T", bf)
	log.Println("type of bloom filter:", typeName)
	log.Printf(
		"fpr = %.2f %%, n = %v, real n = %d / %f%% = %d\nm = %d, k = %d, hash = [%s]\n",
		testFp*100,
		testN,
		testN, populationRatio*100, realTestN,
		optM,
		optK,
		bf.HashAttr(),
	)

	data := [][]byte{}

	for i := uint(0); i < realTestN; i++ {
		value := []byte(fmt.Sprintf("data something is pretty long long long long long to hash%b", i))
		data = append(data, value)
	}
	log.Printf("prepared %d test elements\n", realTestN)

	var addDuration int64 = 0
	var queryDuration int64 = 0

	for i := range data[:testN] {
		addTime := time.Now()
		bf.Add(data[i])
		addDuration += time.Since(addTime).Microseconds()
	}

	log.Printf("added %d test elements (%d mis)\n", testN, addDuration)

	expectedFalseCount := realTestN - testN
	expectedFalsePerc := float64(expectedFalseCount) * 100 / float64(realTestN)
	falseCount := 0
	tp, fp, tn, fn := 0, 0, 0, 0

	for i := range data {
		queryTime := time.Now()
		ok := bf.Contains(data[i])
		queryDuration += time.Since(queryTime).Microseconds()
		if uint(i) < testN {
			// checking added data
			if ok {
				// added and found -> true positive
				tp++
			} else {
				// added but not found -> false negative
				fn++
				falseCount++
			}
		} else {
			// checking unadded data
			if ok {
				// not added but found -> false positive
				fp++
			} else {
				// not added and not found -> true negative
				tn++
				falseCount++
			}
		}
	}

	log.Printf("queried %d elements (%d mis)\n", len(data), queryDuration)

	falsePerc := float64(falseCount*100.0) / float64(realTestN)

	pos, neg := register.GetBitNums(r)
	loadFactor := float64(testN) * 100 / float64(bf.Cap())
	bitLoadFactor := float64(pos) * 100 / float64(pos+neg)

	log.Printf("checked %d test elements\n", realTestN)

	log.Printf("load factor = %.2f %% (%d / %d) \n", loadFactor, testN, bf.Cap())
	log.Printf("bit load factor = %.2f %% (%d / %d) \n", bitLoadFactor, pos, pos+neg)

	log.Printf("false count: %v (%.2f %%)\n", falseCount, falsePerc)
	log.Printf("expected false count: %v (%.2f %%)\n", expectedFalseCount, expectedFalsePerc)

	executionTime := time.Since(start).Microseconds()
	log.Printf("execution time = (%d mis)\n", executionTime)
	return executionTime
}

func TestClassicBloomCreate(t *testing.T) {
	bf := bloomfilter.NewClassicBFBuilder[uint64]().Build()
	typeName := fmt.Sprintf("%T", bf)
	log.Println("type of bloom filter:", typeName)
}

func TestClassicBloomBasic(t *testing.T) {
	log.Printf("\n\n---------classic bloom filter ---------\n\n")

	testFp := 0.1
	testN := uint(4 * 100000)
	populationRatio := 1 / 30.0

	for _, hashFuncAttr := range hasher.HashFunctionAttributes {
		log.Printf("\n\n---------test for standard---------\n\n")
		testClassicBloomHelperBasic(testFp, testN, populationRatio, "standard", hashFuncAttr)
		// log.Printf("\n\n---------test for extended double hashing---------\n\n")
		// testClassicBloomHelperBasic(testFp, testN, populationRatio, "extended-double-hashing", hashFuncAttr)
		// log.Printf("\n\n---------test for kirsch-mitzenmacher---------\n\n")
		// testClassicBloomHelperBasic(testFp, testN, populationRatio, "kirsch-mitzenmacher", hashFuncAttr)

	}
}
