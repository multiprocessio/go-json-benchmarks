#!/usr/bin/env bash

set -e

curl -LO https://s3.amazonaws.com/nyc-tlc/trip+data/yellow_tripdata_2021-04.csv
dsq yellow_tripdata_2021-04.csv > yellow_tripdata_2021-04.json
