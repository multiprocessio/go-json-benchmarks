#!/usr/bin/env bash

set -e

go build -o main

encoders="stdlib nosort stream"
samples="long wide"
times="5"

echo "sample,encoder,time"

for sample in $samples; do
    for encoder in $encoders; do
	./main --in $sample.json --out $sample-$encoder.json --encoder $encoder --ntimes $times
    done
done
