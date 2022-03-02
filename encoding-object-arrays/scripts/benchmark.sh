#!/usr/bin/env bash

set -e

go build -o main

encoders="stdlib,goccy_go-json,nosort,nosort+goccy_go-json,stream"
samples="long wide taxi"
times="10"

echo "sample,encoder,time"

for sample in $samples; do
    ./main --in $sample --encoders $encoders --ntimes $times
done
