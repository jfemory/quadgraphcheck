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
)

//critMatrix is a maxCriticalHeight x maxCriticalCycle indexed matrix of arrays of ints where the ints represent the constants associated with a given preperiodic portrait.
type critMatrix [][][]int

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
	fmt.Println(buildPreperiodicPortrait(13, 2))
	fmt.Println(initializeCritMatrix(41))
	/*
		//writer logic starts here
		file, err := os.Create("output/preperiodicPortraitStats.csv")
		checkError("Cannot Create File. ", err)
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()

		//initialize header of output csv file
		writer.Write([]string{"p", "hAvg", "hMax", "nAvg", "nMax", "tAvg", "tMax", "singletonRatio", "nonsingletonClasses"})
		primeChan := make(chan int)

		portChan := make(chan []funcGraph)
		critCycleCheckRootP(17, 8, portChan, &wg)

		//start prime parser
		//go parsePrimeListCSV(primeChan)

		//var waitToScore sync.WaitGroup

			for {
				portChan := make(chan []funcGraph)
				p := <-primeChan
				buildPrimePortrait(p, portChan, &waitToScore)
				waitToScore.Wait()
				go scorePrimePortrait(p, portChan, writer)

			}
	*/
}

//buildPreperiodicPortrait takes a prime and a constant, returning a funcGraph with constant, critHeight, and critCycleLength filled in.
func buildPreperiodicPortrait(p, c int) funcGraph {
	critHeight, critCycleLength := critHeightAndCycle(p, c)
	return funcGraph{[]int{c}, critHeight, critCycleLength}
}

//critHeightAndCycle takes a prime and a constant, returning the critical point height and critial cycle length.
func critHeightAndCycle(p, c int) (int, int) {
	cycleSlice := []int{0}
	for i := 0; i < p; i++ {
		new := dynamicOperator(p, c, cycleSlice[i])
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

//critMatrixWriter

//initializeCritMatrix taks arguments of an estimated maxH and maxN and initialzes the matrix.
func initializeCritMatrix(p int) critMatrix {
	// * let's start with 300*ln(p) as a guesstimate. Will revise with better data. Examine curve and rewrite.
	upperBound := 300 * int(math.Floor(math.Log(float64(p))))
	matrix := make(critMatrix, upperBound)
	for i := range matrix {
		matrix[i] = make([][]int, upperBound)
	}
	return matrix
	//TODO: Add error
}

//dynamicOperator takes a prime, a constant, and a value, returning the next step in the dynamical system
func dynamicOperator(p int, c int, value int) int {
	return (((value * value) + c) % p)
}

//**********************Old Stuff, refactor or delete***********************************

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
	//fmt.Println(out.p)  //output text for keep alive check
	writeIt([]string{strconv.Itoa(out.p), strconv.FormatFloat(out.hAvg, 'f', -1, 64), strconv.Itoa(out.hMax), strconv.FormatFloat(out.nAvg, 'f', -1, 64), strconv.Itoa(out.nMax), strconv.FormatFloat(out.tAvg, 'f', -1, 64), strconv.Itoa(out.tMax), strconv.FormatFloat(out.singletonRatio, 'f', -1, 64), strconv.Itoa(out.nonsingletonClasses)}, writer)
}

//buildPrimePortrait builds an array of funcGrapheriodic portraits from channel of funcGrapheriodic portraits, as they come in. It returns this array.
func buildPrimePortrait(p int, portChan chan<- []funcGraph, waitToScore *sync.WaitGroup) {
	waitToScore.Add(1)
	primePortrait := make([]funcGraph, 0)
	portrait := make(chan funcGraph)
	var wg sync.WaitGroup
	for i := 1; i < p; i++ {
		go critCycleCheckRootP(p, i, portrait, &wg)
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

//***broken***, need to fix before using - delete if possible if refactored
func critCycleCheckRootP(p int, c int, portrait chan<- funcGraph, wg *sync.WaitGroup) {
	wg.Add(1)
	steps := intSqrt(p)
	fmt.Println("Step size is ", steps)
	cycleSlice := []int{0}
	for i := 0; i < steps+1; i++ {
		new := buildDynamicSlice(p, c, cycleSlice[len(cycleSlice)-1], steps)
		cycleSlice = append(cycleSlice, new...)
		fmt.Println(cycleSlice)
		for j := 0; j < len(new); j++ {
			for k := 0; k < len(cycleSlice); k++ {
				if new[j] == cycleSlice[k] {
					portrait <- funcGraph{[]int{c}, (len(cycleSlice) - k), k}
					fmt.Println("********************")
					wg.Done()
					return
				}
			}
		}
	}
}

//buildDynamicSlice takes a prime, a constant, and a starting value, returning a slice of values from the starting point (not inclusive) for a given number of steps.
func buildDynamicSlice(p, c, startValue, steps int) []int {
	output := []int{dynamicOperator(p, c, startValue)}
	for i := 0; i < steps; i++ {
		output = append(output, dynamicOperator(p, c, output[i]))
	}
	return output
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

//intSqrt takes an integer, takes the square root, and returns the floor function of the result
func intSqrt(input int) int {
	return int(math.Floor(math.Sqrt(float64(input))))
}
