package main

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nnurry/probabilistics/v2/membership/bloomfilter"
	"github.com/nnurry/probabilistics/v2/utilities/hasher"
	"github.com/nnurry/probabilistics/v2/utilities/register"
)

type BenchmarkTime struct {
	ExecutionTime int `json:"execution"`
	AddTime       int `json:"add"`
	QueryTime     int `json:"query"`
}

type BenchmarkAccuracy struct {
	FalseCount struct {
		FalseCount         int `json:"got"`
		ExpectedFalseCount int `json:"expected"`
	} `json:"false_count"`

	FalsePercentage struct {
		FalsePerc         float64 `json:"got"`
		ExpectedFalsePerc float64 `json:"expected"`
	} `json:"false_perc"`

	FalseDiff struct {
		CountDiff int     `json:"count"`
		PercDiff  float64 `json:"perc"`
	} `json:"false_diff"`
}

type BenchmarkResult struct {
	Time                BenchmarkTime     `json:"time"`
	Accuracy            BenchmarkAccuracy `json:"accuracy"`
	BloomTestParameters `json:"test_parameters"`

	BloomN int `json:"bloom_n"`
	RealN  int `json:"real_n"`

	NumHash int `json:"k"`
}

func truncate(val float64, decimal int) float64 {
	if decimal <= 0 {
		decimal = 0
	}
	temp := math.Pow10(decimal)
	return float64(int(val*temp)) / temp
}

func testApp(args []string) {
	var err error

	if len(args) < 1 {
		log.Fatal("don't know what to test\n")
	}

	if args[0] == "generate_parameters" {
		args = args[1:]
		params := &BloomTestParameters{
			FilterType:     "classic",
			Fpr:            0.1,
			GenerateMethod: "standard",
			HashFuncAttr:   hasher.HashFunctionAttributes[0],
		}
		filePath := "./parameters.json"

		if len(args) > 0 {
			params.FilterType = args[0]
			args = args[1:]
		}

		if len(args) > 0 {
			fpr, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				log.Fatal(err)
			}
			params.Fpr = fpr
			args = args[1:]
		}

		if len(args) > 0 {
			params.GenerateMethod = args[0]
			args = args[1:]
		}

		if len(args) > 0 {
			params.HashFuncAttr.HashFamily = args[0]
			args = args[1:]
		}

		if len(args) > 0 {
			platformBit, err := strconv.Atoi(args[0])
			if err != nil {
				log.Fatal(err)
			}

			params.HashFuncAttr.PlatformBit = uint(platformBit)

			args = args[1:]
		}

		if len(args) > 0 {
			outputBit, err := strconv.Atoi(args[0])
			if err != nil {
				log.Fatal(err)
			}

			params.HashFuncAttr.OutputBit = uint(outputBit)
			args = args[1:]
		}

		if len(args) > 0 {
			filePath = args[0]
		}

		generateTestParameters(params, filePath)

	} else if args[0] == "generate_dataset" {
		args = args[1:]
		bloomElements := uint(400000)
		populationRatio := 30.0
		generatorType := "uuid"
		generator := gofakeit.GlobalFaker.UUID
		filePath := "./dataset.json"
		if len(args) > 0 {
			_bloomElements, err := strconv.Atoi(args[0])
			if err != nil {
				log.Fatal(err)
			}
			bloomElements = uint(_bloomElements)
			args = args[1:]
		}
		if len(args) > 0 {
			populationRatio, err = strconv.ParseFloat(args[0], 64)
			if err != nil {
				log.Fatal(err)
			}
			args = args[1:]
		}
		if len(args) > 0 {
			generatorType = args[0]
			switch generatorType {
			case "full_name":
				generator = gofakeit.GlobalFaker.Name
			case "phone_number":
				generator = gofakeit.GlobalFaker.PhoneFormatted
			default:
				generatorType = "uuid"
			}
			args = args[1:]
		}
		if len(args) > 0 {
			filePath = args[0]
		}
		log.Println(
			"generating data with following parameters",
			bloomElements,
			populationRatio,
			generatorType,
			filePath,
		)

		generateDataset(bloomElements, 1/populationRatio, generator, filePath)

	} else if args[0] == "benchmark" {
		args = args[1:]
		paramPath, datasetPath, outputPath := "./parameters.json", "./dataset.json", "output.json"
		if len(args) == 3 {
			paramPath, datasetPath, outputPath = args[0], args[1], args[2]
		}
		params, dataset := initBloomTest(paramPath, datasetPath)
		log.Println("benchmarking")

		start := time.Now() // Start the timer

		testFp := params.Fpr
		testN := dataset.BloomN
		realTestN := dataset.RealN

		optM, optK := bloomfilter.ClassicBFEstimateParams(testFp, testN)
		builder := bloomfilter.NewClassicBFBuilder[uint64]()

		r, _ := register.NewRegister(optM, 1)

		builder = builder.
			SetCap(optM).
			SetHashNum(optK).
			SetRegister(r.(*register.BitRegister)).
			SetHashGenerator(
				params.HashFuncAttr.HashFamily,
				params.HashFuncAttr.PlatformBit,
				params.HashFuncAttr.OutputBit,
				params.GenerateMethod,
			)

		bf := builder.Build()

		log.Println("parameters:", bf, bf.HashAttr())

		var addDuration int64 = 0
		var queryDuration int64 = 0

		for i := range dataset.BloomKeys {
			bloomKey := []byte(dataset.BloomKeys[i])
			addTime := time.Now()
			bf.Add(bloomKey)
			addDuration += time.Since(addTime).Microseconds()
		}

		log.Printf("added %d test elements (%d mis)\n", testN, addDuration)

		expectedFalseCount := realTestN - testN
		expectedFalsePerc := float64(expectedFalseCount) * 100 / float64(realTestN)
		falseCount := 0
		tp, fp, tn, fn := 0, 0, 0, 0

		for i := range dataset.DatasetKeys {
			datasetKey := []byte(dataset.DatasetKeys[i])
			queryTime := time.Now()
			ok := bf.Contains(datasetKey)
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

		log.Printf("queried %d elements (%d mis)\n", realTestN, queryDuration)

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

		result := &BenchmarkResult{
			Time: BenchmarkTime{
				ExecutionTime: int(executionTime),
				QueryTime:     int(queryDuration),
				AddTime:       int(addDuration),
			},
			Accuracy: BenchmarkAccuracy{
				FalseCount: struct {
					FalseCount         int "json:\"got\""
					ExpectedFalseCount int "json:\"expected\""
				}{FalseCount: int(falseCount), ExpectedFalseCount: int(expectedFalseCount)},
				FalsePercentage: struct {
					FalsePerc         float64 "json:\"got\""
					ExpectedFalsePerc float64 "json:\"expected\""
				}{FalsePerc: truncate(falsePerc, 2), ExpectedFalsePerc: truncate(expectedFalsePerc, 2)},
				FalseDiff: struct {
					CountDiff int     "json:\"count\""
					PercDiff  float64 "json:\"perc\""
				}{
					CountDiff: int(falseCount) - int(expectedFalseCount),
					PercDiff:  truncate(falsePerc-expectedFalsePerc, 2),
				},
			},
			BloomN:              int(dataset.BloomN),
			RealN:               int(dataset.RealN),
			NumHash:             int(optK),
			BloomTestParameters: *params,
		}

		data, _ := json.MarshalIndent(result, "", "\t")
		os.WriteFile(outputPath, data, 0777)

	} else {
		log.Fatal("Unrecognizable command\n")
	}
}
