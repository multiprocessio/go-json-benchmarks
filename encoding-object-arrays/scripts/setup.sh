#!/usr/bin/env bash

set -e

curl -LO https://s3.amazonaws.com/nyc-tlc/trip+data/yellow_tripdata_2021-04.csv
dsq yellow_tripdata_2021-04.csv > taxi.json

go install github.com/multiprocessio/fakegen@latest
fakegen --rows 5000 --cols 1000 > wide.json
fakegen --rows 500000 --cols 10 > long.json
