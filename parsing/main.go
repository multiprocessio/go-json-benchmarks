package main

import (
	"io"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	goccy_json "github.com/goccy/go-json"
	sonic_decoder "github.com/bytedance/sonic/decoder"
	jsoniter "github.com/json-iterator/go"
)

func goccy_jsonDecoder(r io.Reader, obj *interface{}) error {
	decoder := goccy_json.NewDecoder(r)
	return decoder.Decode(obj)
}

func stdlibDecoder(r io.Reader, obj *interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(obj)
}

func jsoniterDecoder(r io.Reader, obj *interface{}) error {
	decoder := jsoniter.NewDecoder(r)
	return decoder.Decode(obj)
}

func sonicDecoder(r io.Reader, obj *interface{}) error {
	bs, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	decoder := sonic_decoder.NewDecoder(string(bs))
	return decoder.Decode(obj)
}

func main() {
	var in string
	nTimes := 1
	decoders := []func(io.Reader, *interface{}) error{stdlibDecoder}
	decoderArgs := []string{"stdlib"}

	for i, arg := range os.Args {
		if arg == "--in" {
			in = os.Args[i+1]
			i += 1
			continue
		}

		if arg == "--ntimes" {
			var err error
			nTimes, err = strconv.Atoi(os.Args[i+1])
			if err != nil {
				panic(err)
			}

			i += 1
			continue
		}

		if arg == "--decoders" {
			decoderArgs = strings.Split(os.Args[i+1], ",")
			decoders = nil
			for _, a := range decoderArgs {
				switch a {
				case "stdlib":
					decoders = append(decoders, stdlibDecoder)
				case "goccy":
					decoders = append(decoders, goccy_jsonDecoder)
				case "sonic":
					decoders = append(decoders, sonicDecoder)
				case "jsoniter":
					decoders = append(decoders, jsoniterDecoder)
				}
			}
			i += 1
			continue
		}
	}

	for i, decoder := range decoders {
		decoderArg := decoderArgs[i]

		for i := 0; i < nTimes; i++ {
			fr, err := os.Open(in + ".json")
			if err != nil {
				panic(err)
			}

			var o interface{}
			t1 := time.Now()
			err = decoder(fr, &o)
			t2 := time.Now()
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s,%s,%s\n", in, decoderArg, t2.Sub(t1))
			runtime.GC()

			err = fr.Close()
			if err != nil{
				panic(err)
			}
		}
	}
}
