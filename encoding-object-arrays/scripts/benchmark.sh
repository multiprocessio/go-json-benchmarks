#!/usr/bin/env bash

set -e

go build -o main

rm runlog || echo "No runlog"

encoders="stdlib nosort stream"
samples="long wide"
times="1 2 3 4 5"

echo "encoder,sample,time"

for sample in $samples; do
    for encoder in $encoders; do
        for time in $times; do
	    res="$( TIMEFORMAT=%R; time ( ./main --in $sample.json --out $sample-$encoder.json --encoder $encoder >> runlog 2>&1 ) 2>&1 1>/dev/null )"
    	    echo "$encoder,$sample,$res"
        done
    done
done
