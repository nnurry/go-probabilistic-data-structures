package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nnurry/probabilistics/v2/utilities/hasher"
)

type BloomTestParameters struct {
	FilterType     string               `json:"filter_type"`
	Fpr            float64              `json:"fpr"`
	GenerateMethod string               `json:"generate_method"`
	HashFuncAttr   hasher.HashAttribute `json:"hash_func_attr"`
}

type BloomDataset struct {
	RealN       uint     `json:"real_n"`
	BloomN      uint     `json:"bloom_n"`
	DatasetKeys []string `json:"dataset_keys"`
	BloomKeys   []string `json:"bloom_keys"`
}

func readContent(path string) ([]byte, error) {
	var err error
	var file *os.File
	var content []byte

	file, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err = io.ReadAll(file)
	return content, err
}

func parseBloomTestParameters(bloomTestParametersPath string) *BloomTestParameters {
	errStr := "can't open file to parse test parameters: %s\n"
	content, err := readContent(bloomTestParametersPath)
	if err != nil {
		log.Fatalf(errStr, err)
	}

	bloomTestParameters := &BloomTestParameters{}
	err = json.Unmarshal(content, bloomTestParameters)

	if err != nil {
		log.Fatal(errStr)
	}

	return bloomTestParameters

}
func parseBloomDataset(bloomDatasetPath string) *BloomDataset {
	errStr := "can't open file to parse test parameters: %s\n"
	content, err := readContent(bloomDatasetPath)
	if err != nil {
		log.Fatalf(errStr, err)
	}

	bloomDataset := &BloomDataset{}
	err = json.Unmarshal(content, bloomDataset)

	if err != nil {
		log.Fatal(errStr)
	}

	return bloomDataset
}

func initBloomTest(bloomTestParametersPath string, datasetPath string) (*BloomTestParameters, *BloomDataset) {
	return parseBloomTestParameters(bloomTestParametersPath), parseBloomDataset(datasetPath)
}

func generateTestParameters(testParameters *BloomTestParameters, filePath string) {
	errStr := "can't generate test parameters file: %s\n"

	data, err := json.MarshalIndent(testParameters, "", "\t")
	if err != nil {
		log.Fatalf(errStr, err)
	}

	if err = os.WriteFile(filePath, data, 0777); err != nil {
		log.Fatalf(errStr, err)
	}
}

func generateDataset(bloomElements uint, populationRatio float64, generator func() string, filePath string) {
	errStr := "can't generate dataset file: %s\n"

	testN := bloomElements
	realTestN := uint(float64(testN) / float64(populationRatio))

	bloomKeys := []string{}
	datasetKeys := []string{}

	for i := uint(0); i < realTestN; i++ {
		generatedValue := generator()
		value := fmt.Sprintf("%s-%d", generatedValue, i%testN)
		if i < testN {
			bloomKeys = append(bloomKeys, value)
		}
		datasetKeys = append(datasetKeys, value)
	}

	dataset := &BloomDataset{
		BloomN:      testN,
		RealN:       realTestN,
		BloomKeys:   bloomKeys,
		DatasetKeys: datasetKeys,
	}

	data, err := json.MarshalIndent(dataset, "", "\t")
	if err != nil {
		log.Fatalf(errStr, err)
	}

	if err = os.WriteFile(filePath, data, 0777); err != nil {
		log.Fatalf(errStr, err)
	}
}
