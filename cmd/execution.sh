#!/bin/bash

# Default values
DATASET_SIZE=${1:-400000}
POPULATION_RATIO=${2:-1.32}
DATASET_TYPE=${3:-uuid}
BLOOM_TYPE=${4:-classic}
BLOOM_PARAMETER=${5:-"0.025 standard"}
OUTPUT_PREFIX=${6:-"./output"}

# Clean up previous outputs
rm output/* parameters/* dataset/* 2>/dev/null
mkdir -p parameters dataset output

# Generate dataset
DATASET_FILE="./dataset/${DATASET_SIZE}-${POPULATION_RATIO}-${DATASET_TYPE}.json"
go run . test generate_dataset "$DATASET_SIZE" "${POPULATION_RATIO}" "$DATASET_TYPE" "$DATASET_FILE"

# Generate parameters
PARAMS=(
    "murmur3Hash128Default 64 128"
    "murmur3Hash128Spaolacci 64 128"
    "murmur3Hash64Spaolacci 64 64"
    "murmur3Hash256Bnb 64 256"
    "xxHashCespare 64 64"
    "xxHashOneOfOne 64 64"
)

for param in "${PARAMS[@]}"; do
    go run . test generate_parameters "$BLOOM_TYPE" $BLOOM_PARAMETER $param "./parameters/${BLOOM_TYPE}-${BLOOM_PARAMETER// /-}-$param-parameters.json"
done

# Benchmarking
for param in "${PARAMS[@]}"; do
    go run . test benchmark "./parameters/${BLOOM_TYPE}-${BLOOM_PARAMETER// /-}-$param-parameters.json" "$DATASET_FILE" "./output/${BLOOM_TYPE}-${BLOOM_PARAMETER// /-}-$param-benchmark-result.json"
done
