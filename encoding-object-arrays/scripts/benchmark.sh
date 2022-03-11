#!/usr/bin/env bash

set -e

go build -o main

encoders="stdlib,segment,goccy,nosort,jsoniter,sonic,nosort_segment,nosort_goccy,nosort_jsoniter,nosort_sonic"
samples="long wide taxi"
times="2"

echo "sample,encoder,time"

for sample in $samples; do
    ./main --in $sample --encoders $encoders --ntimes $times
done
