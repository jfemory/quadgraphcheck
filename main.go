package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

//funcGraph is a struct representing a functional graph.
type funcGraph struct {
	constant        []int
	critHeight      int
	critCycleLength int
	//components      []block
	//tempList	[][]int //edgeset is temporary and decrements as components are built.
}
type outputData struct {
	p                   int
	hAvg                float64
	hMax                int
	nAvg                float64
	nMax                int
	tAvg                float64
	tMax                int
	singletonRatio      float64
	nonsingletonClasses int
}

func main() {
	//writer logic starts here
	file, err := os.Create("output/preperiodicPortraitStats.csv")
	checkError("Cannot Create File. ", err)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	//initialize header of output csv file
	writer.Write([]string{"p", "hAvg", "hMax", "nAvg", "nMax", "tAvg", "tMax", "singletonRatio", "nonsingletonClasses"})
	primeChan := make(chan int)

	go parsePrimeListCSV(primeChan)

	var waitToScore sync.WaitGroup

	for {
		portChan := make(chan []funcGraph)
		p := <-primeChan
		go buildPrimePortrait(p, portChan, &waitToScore)
		go scorePrimePortrait(p, portChan, writer)
		waitToScore.Wait()
	}
}

//scorePrimePortrait takes a prime portrait slice, port, and puts outputData onto the outData channel
func scorePrimePortrait(p int, portChan <-chan []funcGraph, writer *csv.Writer) {
	port := <-portChan
	var out outputData
	hMax := 0
	hSum := 0
	nMax := 1
	nSum := 1
	tSum := 0
	tCount := 0
	tMax := 0
	singletonCount := 1
	out.p = p

	for i := 0; i < len(port); i++ {
		//fmt.Println("a")
		//set coefficient for sums
		coeff := len(port[i].constant)
		//increment x_sum and x_count
		hSum = hSum + (coeff * port[i].critHeight)
		nSum = nSum + (coeff * port[i].critCycleLength)
		if len(port[i].constant) != 1 {
			tSum = tSum + len(port[i].constant)
			tCount++
		} else {
			singletonCount++
		}
		//update x_max
		if hMax < port[i].critHeight {
			hMax = port[i].critHeight
		}
		if nMax < port[i].critCycleLength {
			nMax = port[i].critCycleLength
		}
		if tMax < len(port[i].constant) {
			tMax = len(port[i].constant)
		}
	}
	out.hAvg = float64(hSum) / float64(p)
	out.hMax = hMax
	out.nAvg = float64(nSum) / float64(p)
	out.nMax = nMax
	if tCount == 0 {
		out.tAvg = 1
	} else {
		out.tAvg = float64(tSum) / float64(tCount)
	}
	out.tMax = tMax
	out.singletonRatio = float64(singletonCount) / float64(p)
	out.nonsingletonClasses = tCount
	fmt.Println(out.p)
	writeIt([]string{strconv.Itoa(out.p), strconv.FormatFloat(out.hAvg, 'f', -1, 64), strconv.Itoa(out.hMax), strconv.FormatFloat(out.nAvg, 'f', -1, 64), strconv.Itoa(out.nMax), strconv.FormatFloat(out.tAvg, 'f', -1, 64), strconv.Itoa(out.tMax), strconv.FormatFloat(out.singletonRatio, 'f', -1, 64), strconv.Itoa(out.nonsingletonClasses)}, writer)
}

//buildPrimePortrait builds an array of funcGrapheriodic portraits from channel of funcGrapheriodic portraits, as they come in. It returns this array.
func buildPrimePortrait(p int, portChan chan<- []funcGraph, waitToScore *sync.WaitGroup) {
	waitToScore.Add(1)
	primePortrait := make([]funcGraph, 0)
	portrait := make(chan funcGraph)
	var wg sync.WaitGroup
	for i := 1; i < p; i++ {
		go funcGrapheriod(p, i, portrait, &wg)
	}
	for i := 1; i < p; i++ {
		flag := false
		new := <-portrait
		for j := 0; j < len(primePortrait); j++ {
			if primePortrait[j].critCycleLength == new.critCycleLength && primePortrait[j].critHeight == new.critHeight {
				flag = true
				primePortrait[j].constant = append(primePortrait[j].constant, new.constant[0])
				break
			}

		}
		if flag == false {
			primePortrait = append(primePortrait, new)
		}
	}
	wg.Wait()
	waitToScore.Done()
	close(portrait)
	portChan <- primePortrait
}

//funcGrapheriod takes a prime p and a constant c, putting a funcGraph onto the portrait chan. Run as a go routine.
func funcGrapheriod(p int, c int, portrait chan<- funcGraph, wg *sync.WaitGroup) {
	wg.Add(1)
	//fmt.Println(c)
	cycleCheck := make([]int, 0)
	cycleCheck = append(cycleCheck, 0)
	var new int
	for i := 0; i < p; i++ {
		new = (cycleCheck[i]*cycleCheck[i] + c) % p
		for j := 0; j < len(cycleCheck); j++ {
			if new == cycleCheck[j] {
				portrait <- funcGraph{[]int{c}, (len(cycleCheck) - j), j}
				wg.Done()
				return
			}
		}
		cycleCheck = append(cycleCheck, new)
	}
}

//parsePrimeList takes a CSV of primes and pushes them one by one onto primeChan
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

//parsePrimeList takes a list of primes and pushes them one by one onto primeChan
func parsePrimeList(primeChan chan int) {
	defer close(primeChan)
	//open file logic
	openFile, err := os.Open("list/list.prime")
	checkError("Failed to open prime list file. ", err)
	defer openFile.Close()

	scanner := bufio.NewScanner(openFile)
	for scanner.Scan() {
		stringSlice := (strings.Split(strings.Trim(scanner.Text(), "[]"), " "))
		for i := 0; i < len(stringSlice); i++ {
			prime, _ := strconv.Atoi(stringSlice[i])
			primeChan <- prime
		}
	}
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func writeIt(data []string, writer *csv.Writer) {
	err := writer.Write(data)
	checkError("Write to file failed. ", err)
}
