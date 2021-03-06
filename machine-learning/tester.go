package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		panic("usage: command model_name model_version data_file")
	}
	model, version := os.Args[1], os.Args[2]
	lines := readDataFile(os.Args[3])
	serviceURL := fmt.Sprintf("http://localhost:8501/v1/models/%s/versions/%s:predict", model, version)

	body, err := json.Marshal(&Req{
		SignatureName: "raw",
		Instances: []Instance{
			Instance{Raw: lineToFloats(<-lines)},
			Instance{Raw: lineToFloats(<-lines)},
			Instance{Raw: lineToFloats(<-lines)},
			Instance{Raw: lineToFloats(<-lines)},
		},
	})

	if err != nil {
		log.Fatal("json.Marshal error", err)
	}

	start := time.Now()
	resp, err := http.Post(serviceURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal("http post error", err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("ioutil.ReadAll error", err)
	}
	elapsed := time.Since(start)

	println(string(respBody))
	fmt.Printf("elapsed: %s\n", elapsed.String())
}

func lineToFloats(line string) []float64 {
	vals := strings.Split(line, ",")
	floats := make([]float64, len(vals))
	for i, s := range vals {
		v, _ := strconv.ParseFloat(s, 64)
		floats[i] = v
	}
	return floats[0 : len(floats)-2]
}

func readDataFile(path string) <-chan string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)

	lines := make(chan string)
	go func() {
		defer file.Close()
		defer close(lines)
		for scanner.Scan() {
			lines <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}()

	return lines
}

type Instance struct {
	Raw []float64 `json:"raw"`
}

type Req struct {
	SignatureName string     `json:"signature_name"`
	Instances     []Instance `json:"instances"`
}
