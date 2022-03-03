package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	goccy_json "github.com/goccy/go-json"
)

func nosortEncoder(out *os.File, obj interface{}, marshalFn func(o interface{}) ([]byte, error)) error {
	a, ok := obj.([]interface{})
	// Fall back to normal encoder
	if !ok {
		log.Println("Falling back to stdlib")
		return stdlibEncoder(out, obj)
	}

	bo := bytes.NewBuffer(nil)
	_, err := bo.Write([]byte("["))
	if err != nil {
		return err
	}

	quotedColumns := map[string][]byte{}

	for i, row := range a {
		// Write a comma before the current object
		if i > 0 {
			_, err = bo.Write([]byte(",\n"))
			if err != nil {
				return err
			}
		}

		r, ok := row.(map[string]interface{})
		if !ok {
			log.Println("Falling back to stdlib")
			bs, err := marshalFn(row)
			if err != nil {
				return err
			}

			_, err = bo.Write(bs)
			if err != nil {
				return err
			}
			continue
		}

		_, err := bo.Write([]byte("{"))
		if err != nil {
			return err
		}

		j := -1
		for col, val := range r {
			j += 1

			// Write a comma before the current key-value
			if j > 0 {
				_, err = bo.Write([]byte(","))
				if err != nil {
					return err
				}
			}

			quoted := quotedColumns[col]
			if quoted == nil {
				quoted = []byte(strconv.QuoteToASCII(col) + ":")
				quotedColumns[col] = quoted
			}
			_, err = bo.Write(quoted)
			if err != nil {
				return err
			}

			bs, err := marshalFn(val)
			if err != nil {
				return err
			}

			_, err = bo.Write(bs)
			if err != nil {
				return err
			}
		}

		_, err = bo.Write([]byte("}"))
		if err != nil {
			return err
		}
	}

	_, err = bo.Write([]byte("]"))

	for bo.Len() > 0 {
		_, err := bo.WriteTo(out)
		if err != nil {
			return err
		}
	}

	return err
}

func stdlibEncoder(out *os.File, obj interface{}) error {
	encoder := json.NewEncoder(out)
	return encoder.Encode(obj)
}

func goccy_jsonEncoder(out *os.File, obj interface{}) error {
	encoder := goccy_json.NewEncoder(out)
	return encoder.Encode(obj)
}

func main() {
	var in string
	nTimes := 1
	encoders := []func(*os.File, interface{}) error{stdlibEncoder}
	encoderArgs := []string{"stdlib"}

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

		if arg == "--encoders" {
			encoderArgs = strings.Split(os.Args[i+1], ",")
			encoders = nil
			for _, a := range encoderArgs {
				switch a {
				case "stdlib":
					encoders = append(encoders, stdlibEncoder)
				case "nosort":
					encoders = append(encoders, func(out *os.File, o interface{}) error {
						return nosortEncoder(out, o, json.Marshal)
					})
				case "nosort_goccy":
					encoders = append(encoders, func(out *os.File, o interface{}) error {
						return nosortEncoder(out, o, goccy_json.Marshal)
					})
				case "goccy_json":
					encoders = append(encoders, goccy_jsonEncoder)
				}
			}
			i += 1
			continue
		}
	}

	fr, err := os.Open(in + ".json")
	if err != nil {
		panic(err)
	}
	defer fr.Close()

	decoder := json.NewDecoder(fr)
	var o interface{}
	err = decoder.Decode(&o)
	if err != nil {
		panic(err)
	}

	fmt.Println("sample,encoder,time")
	for i, encoder := range encoders {
		encoderArg := encoderArgs[i]

		for i := 0; i < nTimes; i++ {
			fw, err := os.OpenFile(in+"-"+encoderArg+".json", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
			if err != nil {
				panic(err)
			}

			t1 := time.Now()
			err = encoder(fw, o)
			t2 := time.Now()
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s,%s,%s\n", in, encoderArg, t2.Sub(t1))
			runtime.GC()

			err = fw.Close()
			if err != nil{
				panic(err)
			}
		}
	}
}
