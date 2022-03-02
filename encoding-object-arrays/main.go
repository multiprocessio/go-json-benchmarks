package main

import (
	"bufio"
	"encoding/json"
	"runtime"
	"strings"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/pkg/profile"
	goccy_json "github.com/goccy/go-json"
)

var (
	J_COMMA     = []byte(",")
	J_COMMA_NL  = []byte(",\n")
	J_ARR_OPEN  = []byte("[")
	J_ARR_CLOSE = []byte("]")
	J_OBJ_OPEN  = []byte("{")
	J_OBJ_CLOSE = []byte("}")
)

func nosortEncoder(marshal func(interface{}) ([]byte, error)) func (out *os.File, obj interface{}) error {
	return func (out *os.File, obj interface{}) error {
		a, ok := obj.([]interface{})
		// Fall back to normal encoder
		if !ok {
			log.Println("Falling back to stdlib")
			return stdlibEncoder(out, obj)
		}

		bo := bufio.NewWriter(out)
		_, err := bo.Write(J_ARR_OPEN)
		if err != nil {
			return err
		}

		quotedColumns := map[string][]byte{}

		for i, row := range a {
			// Write a comma before the current object
			if i > 0 {
				_, err = bo.Write(J_COMMA_NL)
				if err != nil {
					return err
				}
			}

			r, ok := row.(map[string]interface{})
			if !ok {
				log.Println("Falling back to stdlib")
				bo.Flush()
				err = stdlibEncoder(out, r)
				if err != nil {
					return err
				}
				continue
			}

			_, err := bo.Write(J_OBJ_OPEN)
			if err != nil {
				return err
			}

			j := -1
			for col, val := range r {
				j += 1
				cellBytes, err := marshal(val)
				if err != nil {
					return err
				}

				// Write a comma before the current key-value
				if j > 0 {
					_, err = bo.Write(J_COMMA)
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

				_, err = bo.Write(cellBytes)
				if err != nil {
					return err
				}
			}

			_, err = bo.Write(J_OBJ_CLOSE)
			if err != nil {
				return err
			}
		}

		_, err = bo.Write(J_ARR_CLOSE)
		return err
	}
}

func streamEncoder(out *os.File, obj interface{}) error {
	a, ok := obj.([]interface{})
	// Fall back to normal encoder
	if !ok {
		log.Println("Falling back to stdlib")
		return stdlibEncoder(out, obj)
	}

	bo := bufio.NewWriter(out)
	_, err := bo.Write(J_ARR_OPEN)
	if err != nil {
		return err
	}

	for i, row := range a {
		// Write a comma before the current object
		if i > 0 {
			_, err = bo.Write(J_COMMA_NL)
			if err != nil {
				return err
			}
		}

		obj, err := json.Marshal(row)
		if err != nil {
			return err
		}

		_, err = bo.Write(obj)
		if err != nil {
			return err
		}
	}

	_, err = bo.Write(J_ARR_CLOSE)
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
	var nTimes int = 1
	encoders := []func(*os.File, interface{}) error {stdlibEncoder}
	var encoderArgs []string

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
					encoders = append(encoders,nosortEncoder(json.Marshal))
				case "stream":
					encoders = append(encoders, streamEncoder)
				case "goccy_go-json":
					encoders = append(encoders, goccy_jsonEncoder)
				case "nosort+goccy_go-json":
					encoders = append(encoders, nosortEncoder(goccy_json.Marshal))
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

	for i, encoder := range encoders {
		encoderArg := encoderArgs[i]
		fw, err := os.OpenFile(in + "-" + encoderArg + ".json", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer fw.Close()
		
		for i := 0; i < nTimes; i++ {	
			t1 := time.Now()
			//defer profile.Start().Stop()
			err = encoder(fw, o)
			t2 := time.Now()
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s,%s,%s\n", in, encoderArg, t2.Sub(t1))
			runtime.GC()
		}
	}
}
