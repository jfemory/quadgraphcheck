package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strconv"
)

//parsePrimeListCSV takes a CSV of primes and pushes them one by one onto primeChan. Pass a prime chan to ParsePrimeList.
func parsePrimeListCSV(primeChan chan int) {
	defer close(primeChan)
	//open file logic
	openFile, err := os.Open("list/list.prime")
	checkError("Failed to open prime list file. ", err)
	defer openFile.Close()

	reader := csv.NewReader(bufio.NewReader(openFile))
	for {
		stringSlice, error := reader.Read()
		if error == io.EOF {
			break
		}
		for i := 0; i < len(stringSlice); i++ {
			prime, _ := strconv.Atoi(stringSlice[i])
			primeChan <- prime
		}
	}
}

//TODO: Add file argument and default file fallthrough case.
//TODO: Add command line argument for filename.
