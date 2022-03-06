#!/usr/bin/env bash

set -e

go build -o main

decoders="stdlib,goccy,jsoniter,sonic"
samples="long wide taxi"
times="2"

echo "sample,decoder,time"

for sample in $samples; do
    ./main --in $sample --decoders $decoders --ntimes $times
done
