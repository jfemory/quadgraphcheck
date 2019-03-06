/*#quadgraphcheck
Quadratic Functional Graph Isomorphism Checker

The quadratic functional graph isomorphism checker takes a prime number and checks
to see if any of the associated functional graphs are isomorphic. Uploads will be
coming over the next few days.

Place a list of primes in a folder called "list/list.prime" under the
source directory.*/
package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type critMatrixConstants []int

//critMatrix is a maxCriticalHeight x maxCriticalCycle indexed matrix of arrays of ints where the ints represent the constants associated with a given preperiodic portrait.
type critMatrix [][]critMatrixConstants

//critcritMatrixEntry is passed to the matrix writer function and contains the index as a length 2 array
//and the constant to be written.
type critMatrixEntry struct {
	h        int
	n        int
	constant int
}

//funcGraph is a struct representing a functional graph.
type funcGraph struct {
	constant        []int
	critHeight      int
	critCycleLength int
	//components      []block
	//tempList	[][]int //edgeset is temporary and decrements as components are built.
}

type preperiodicOutputData struct {
	prime               int
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
	var nextPrimeWG sync.WaitGroup
	//writePreperiodicStatsChan := make(chan preperiodicOutputData)
	computePrimeStats(17, &nextPrimeWG)
	/*
		//writer logic starts here
		file, err := os.Create("output/preperiodicPortraitStats.csv")
		checkError("Cannot Create File. ", err)
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()

		//initialize header of output csv file
		writer.Write([]string{"prime", "hAvg", "hMax", "nAvg", "nMax", "tAvg", "tMax", "singletonRatio", "nonsingletonClasses"})
	*/
}

//computePrimeStats takes a prime and builds all the statistics relating to that prime.
//It puts the stats onto various channels that have writers for different output files.
//nextPrimeWG lets the calling function know when to start the next prime computation.
//this should occur after completing critMatrix writing but before critMatrix scoring.
func computePrimeStats(prime int, nextPrimeWG *sync.WaitGroup) {
	nextPrimeWG.Add(prime - 1)
	critMatrixEntryChan := make(chan critMatrixEntry)

	go critMatrixWriter(prime, nextPrimeWG, critMatrixEntryChan)
	for constant := 1; constant < prime; constant++ {
		go buildPreperiodicPortrait(prime, constant, nextPrimeWG, critMatrixEntryChan)
	}
	nextPrimeWG.Wait()
	close(critMatrixEntryChan)
	time.Sleep(5 * time.Second)
}

//buildPreperiodicPortrait takes a prime and a constant, returning a funcGraph with constant, critHeight, and critCycleLength filled in.
func buildPreperiodicPortrait(prime, constant int, nextPrimeWG *sync.WaitGroup, critMatrixEntryChan chan<- critMatrixEntry) {
	critHeight, critCycleLength := critHeightAndCycle(prime, constant)
	critMatrixEntryChan <- critMatrixEntry{critHeight, critCycleLength, constant}
	nextPrimeWG.Done()
}

//critHeightAndCycle takes a prime and a constant, returning the critical point height and critial cycle length.
func critHeightAndCycle(prime, constant int) (int, int) {
	cycleSlice := []int{0}
	for i := 0; i < prime; i++ {
		new := dynamicOperator(prime, constant, cycleSlice[i])
		for j := 0; j < len(cycleSlice); j++ {
			if new == cycleSlice[j] {
				return j, (len(cycleSlice) - j)
			}
		}
		cycleSlice = append(cycleSlice, new)
	}
	//TODO: put in error checking here
	return -1, -1
}

//critMatrixWriter is run as a goroutine. It takes a channel of matrixEntries, and writes them to the matrix as they come in.
//This funtion also initializes its own matrix.
func critMatrixWriter(prime int, nextPrimeWG *sync.WaitGroup, critMatrixEntryChan <-chan critMatrixEntry) {
	critMatrix := initializeCritMatrix(prime)
	for i := 0; i < prime-1; i++ {
		a := <-critMatrixEntryChan //a is a matrix entry from the channel
		critMatrix[a.h][a.n] = append(critMatrix[a.h][a.n], a.constant)
		fmt.Println(critMatrix)
		fmt.Println(i)
	}
	nextPrimeWG.Wait()
}

//initializeCritMatrix takes a prime and initialzes the matrix that is hopefully big enough.
func initializeCritMatrix(prime int) critMatrix {
	// * let's start with 300*ln(prime) as a guesstimate. Will revise with better data. Examine curve and rewrite.
	//upperBound := 10 * int(math.Floor(math.Log(float64(prime))))
	upperBound := prime
	matrix := make(critMatrix, upperBound)
	for i := range matrix {
		matrix[i] = make([]critMatrixConstants, upperBound)
		for j := range matrix[i] {
			matrix[i][j] = make(critMatrixConstants, 0)
		}
	}
	return matrix
	//TODO: Add error
}

//dynamicOperator takes a prime, a constant, and a value, returning the next step in the dynamical system
func dynamicOperator(prime int, constant int, value int) int {
	return (((value * value) + constant) % prime)
}

//**********************Old Stuff, refactor or delete***********************************

//scorePrimePortrait takes a prime portrait slice, port, and puts outputData onto the outData channel
func scorePrimePortrait(prime int, portChan <-chan []funcGraph, writer *csv.Writer) {
	port := <-portChan
	var out preperiodicOutputData
	hMax := 0
	hSum := 0
	nMax := 1
	nSum := 1
	tSum := 0
	tCount := 0
	tMax := 0
	singletonCount := 1
	out.prime = prime

	for i := 0; i < len(port); i++ {
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
	out.hAvg = float64(hSum) / float64(prime)
	out.hMax = hMax
	out.nAvg = float64(nSum) / float64(prime)
	out.nMax = nMax
	if tCount == 0 {
		out.tAvg = 1
	} else {
		out.tAvg = float64(tSum) / float64(tCount)
	}
	out.tMax = tMax
	out.singletonRatio = float64(singletonCount) / float64(prime)
	out.nonsingletonClasses = tCount
	//fmt.Println(out.prime)  //output text for keep alive check
	writeIt([]string{strconv.Itoa(out.prime), strconv.FormatFloat(out.hAvg, 'f', -1, 64), strconv.Itoa(out.hMax), strconv.FormatFloat(out.nAvg, 'f', -1, 64), strconv.Itoa(out.nMax), strconv.FormatFloat(out.tAvg, 'f', -1, 64), strconv.Itoa(out.tMax), strconv.FormatFloat(out.singletonRatio, 'f', -1, 64), strconv.Itoa(out.nonsingletonClasses)}, writer)
}

//buildPrimePortrait builds an array of funcGrapheriodic portraits from channel of funcGrapheriodic portraits, as they come in. It returns this array.
func buildPrimePortrait(prime int, portChan chan<- []funcGraph, waitToScore *sync.WaitGroup) {
	waitToScore.Add(1)
	primePortrait := make([]funcGraph, 0)
	portrait := make(chan funcGraph)
	var wg sync.WaitGroup
	for i := 1; i < prime; i++ {
		go critCycleCheckRootP(prime, i, portrait, &wg)
	}
	for i := 1; i < prime; i++ {
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

//***broken***, need to fix before using - delete if possible if refactored
func critCycleCheckRootP(prime int, constant int, portrait chan<- funcGraph, wg *sync.WaitGroup) {
	wg.Add(1)
	steps := intSqrt(prime)
	fmt.Println("Step size is ", steps)
	cycleSlice := []int{0}
	for i := 0; i < steps+1; i++ {
		new := buildDynamicSlice(prime, constant, cycleSlice[len(cycleSlice)-1], steps)
		cycleSlice = append(cycleSlice, new...)
		fmt.Println(cycleSlice)
		for j := 0; j < len(new); j++ {
			for k := 0; k < len(cycleSlice); k++ {
				if new[j] == cycleSlice[k] {
					portrait <- funcGraph{[]int{constant}, (len(cycleSlice) - k), k}
					fmt.Println("********************")
					wg.Done()
					return
				}
			}
		}
	}
}

//buildDynamicSlice takes a prime, a constant, and a starting value, returning a slice of values from the starting point (not inclusive) for a given number of steps.
func buildDynamicSlice(prime, constant, startValue, steps int) []int {
	output := []int{dynamicOperator(prime, constant, startValue)}
	for i := 0; i < steps; i++ {
		output = append(output, dynamicOperator(prime, constant, output[i]))
	}
	return output
}

//parsePrimeListCSV takes a CSV of primes and pushes them one by one onto primeChan
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

//intSqrt takes an integer, takes the square root, and returns the floor function of the result
func intSqrt(input int) int {
	return int(math.Floor(math.Sqrt(float64(input))))
}
