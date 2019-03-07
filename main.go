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
	//"time"
)

type critMatrixConstants []int

//critMatrix is a maxCriticalHeight x maxCriticalCycle indexed matrix of arrays of ints where the ints represent the constants associated with a given preperiodic portrait.
type critMatrix [][]critMatrixConstants

//critMatrixEntry is passed to the matrix writer function and contains the index as a length 2 array
//and the constant to be written.
type critMatrixEntry struct {
	h        int
	n        int
	constant int
}

//funcGraphPackage is a struct holding a prime and and a slice of funcgraphs to be parsed
type funcGraphPackage struct {
	prime  int
	graphs []funcGraph
}

//funcGraph is a struct representing a functional graph.
type funcGraph struct {
	constant        []int
	critHeight      int
	critCycleLength int
	//components      []block
	//tempList	[][]int //edgeset is temporary and decrements as components are built.
}

//preperiodicCounter is a working struct to keep track
//of important values for computing preperiodic stats
type preperiodicCounter struct {
	hMax  int
	hSum  int
	nMax  int
	nSum  int
	hnMax int
	hnSum int
}

type scraperCounter struct {
	tMax            int
	tSum            int
	singletonSum    int
	nonsingletonSum int
}

type preperiodicOutputData struct {
	prime               int
	hAvg                float64
	hMax                int
	nAvg                float64
	nMax                int
	hnAvg               float64
	hnMax               int
	tAvg                float64
	tMax                int
	singletonRatio      float64
	nonsingletonClasses int
}

func main() {
	//initialize preperiodicStatsWriter
	filePreP, err := os.Create("output/preperiodicPortraitStats.csv")
	checkError("Cannot Create File. ", err)
	defer filePreP.Close()
	preperiodicWriter := csv.NewWriter(filePreP)
	defer preperiodicWriter.Flush()

	//initialize zeroIsoWriter
	fileZeroIso, err := os.Create("output/zeroIsoStats.csv")
	checkError("Cannot Create File. ", err)
	defer fileZeroIso.Close()
	zeroIsoWriter := csv.NewWriter(fileZeroIso)
	defer zeroIsoWriter.Flush()

	//initialize primeChan and populate it with prime numbers
	primeChan := make(chan int)
	go parsePrimeListCSV(primeChan)

	//initialize header of output csv file
	preperiodicWriter.Write([]string{"prime", "hAvg", "hMax", "nAvg", "nMax", "hnAvg", "hnMax", "tAvg", "tMax", "singletonRatio", "nonsingletonClasses"})
	//
	//* Select your primes, here.
	for {
		var nextPrimeWG sync.WaitGroup
		p := <-primeChan
		go computePrimeStats(p, &nextPrimeWG, preperiodicWriter)
		nextPrimeWG.Wait()
	}
}

//computePrimeStats takes a prime and builds all the statistics relating to that prime.
//It puts the stats onto various channels that have writers for different output files.
//nextPrimeWG lets the calling function know when to start the next prime computation.
//this should occur after completing critMatrix writing but before critMatrix scoring.
func computePrimeStats(prime int, nextPrimeWG *sync.WaitGroup, preperiodicWriter *csv.Writer) {
	nextPrimeWG.Add(prime - 1)
	critMatrixEntryChan := make(chan critMatrixEntry)

	go critMatrixWriter(prime, nextPrimeWG, critMatrixEntryChan, preperiodicWriter)
	for constant := 1; constant < prime; constant++ {
		go buildPreperiodicPortrait(prime, constant, nextPrimeWG, critMatrixEntryChan)
	}
	nextPrimeWG.Wait()
	close(critMatrixEntryChan)
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
func critMatrixWriter(prime int, nextPrimeWG *sync.WaitGroup, critMatrixEntryChan <-chan critMatrixEntry, preperiodicWriter *csv.Writer) {
	matrix := initializeCritMatrix(prime)
	counter := preperiodicCounter{0, 0, 0, 1, 1, 1} //initialie counter after c=0 is accounted for
	for i := 0; i < prime-1; i++ {
		a := <-critMatrixEntryChan //a is a matrix entry from the channel
		incrementPreperiodicCounter(&counter, a)
		matrix[a.h][a.n] = append(matrix[a.h][a.n], a.constant)
	}
	out := scoreCritMatrix(prime, matrix, counter)
	fmt.Println(out.prime)
	writeIt([]string{strconv.Itoa(out.prime), strconv.FormatFloat(out.hAvg, 'f', -1, 64), strconv.Itoa(out.hMax), strconv.FormatFloat(out.nAvg, 'f', -1, 64), strconv.Itoa(out.nMax), strconv.FormatFloat(out.hnAvg, 'f', -1, 64), strconv.Itoa(out.hnMax), strconv.FormatFloat(out.tAvg, 'f', -1, 64), strconv.Itoa(out.tMax), strconv.FormatFloat(out.singletonRatio, 'f', -1, 64), strconv.Itoa(out.nonsingletonClasses)}, preperiodicWriter)
	nextPrimeWG.Wait()

}
func incrementPreperiodicCounter(counter *preperiodicCounter, entry critMatrixEntry) {
	//update hMax
	if counter.hMax < entry.h {
		counter.hMax = entry.h
	}
	//update hSum
	counter.hSum = counter.hSum + entry.h
	//update nMax
	if counter.nMax < entry.n {
		counter.nMax = entry.n
	}
	//update nSum
	counter.nSum = counter.nSum + entry.n
	//update hnMax
	if counter.hnMax < (entry.h + entry.n) {
		counter.hnMax = (entry.h + entry.n)
	}
	//update hnSum
	counter.hnSum = counter.hnSum + (entry.h + entry.n)
}

func scoreCritMatrix(prime int, matrix critMatrix, counter preperiodicCounter) preperiodicOutputData {
	//initialize out with prime, hMax, nMax, and hnMax.
	out := preperiodicOutputData{prime, 0, counter.hMax, 0, counter.nMax, 0, counter.hnMax, 0, 0, 0, 0}
	graphs := funcGraphPackage{prime, make([]funcGraph, 0)}
	//need to set hAvg, nAvg, hnAvg, tAvg, tMax, singletonRatio and nonsingletonClasses.
	//set hAvg
	out.hAvg = float64(counter.hSum) / float64(prime)
	out.nAvg = float64(counter.nSum) / float64(prime)
	out.hnAvg = float64(counter.hnSum) / float64(prime)
	out.tAvg, out.tMax, out.singletonRatio, out.nonsingletonClasses, graphs.graphs = matrixScraper(matrix)
	return out
}

func matrixScraper(matrix critMatrix) (float64, int, float64, int, []funcGraph) {
	//also return []funcGraph
	counter := scraperCounter{0, 0, 0, 0}
	graphSlice := make([]funcGraph, 0)
	for h := 0; h < len(matrix); h++ {
		for n := 0; n < len(matrix[h]); n++ {
			if len(matrix[h][n]) == 1 {
				counter.singletonSum++ //increment singletonSum
			} else if len(matrix[h][n]) == 2 {
				counter.nonsingletonSum++ //increment nonsingletonSum
				newT := len(matrix[h][n])
				counter.tSum = counter.tSum + newT //increment tSum
				if counter.tMax < newT {
					counter.tMax = newT //update tMax
				}
				graphSlice = append(graphSlice, funcGraph{matrix[h][n], h, n})
			}
		}
	}
	tAvg := float64(counter.tSum) / float64(counter.nonsingletonSum)
	singletonRatio := float64(counter.singletonSum) / float64(counter.singletonSum+counter.nonsingletonSum)

	return tAvg, counter.tMax, singletonRatio, len(graphSlice), graphSlice
}

//initializeCritMatrix takes a prime and initialzes the matrix that is hopefully big enough.
func initializeCritMatrix(prime int) critMatrix {
	// * let's start with 300*ln(prime) as a guesstimate. Will revise with better data. Examine curve and rewrite.
	//upperBound := 300 * int(math.Floor(math.Log(float64(prime))))
	upperBound := prime
	matrix := make(critMatrix, upperBound)
	for i := range matrix {
		matrix[i] = make([]critMatrixConstants, upperBound)
		for j := range matrix[i] {
			matrix[i][j] = make(critMatrixConstants, 0)
		}
	}
	matrix[0][1] = append(matrix[0][1], 0) //initialize c=0 graph. Always singleton.
	return matrix
	//TODO: Add error
}

//dynamicOperator takes a prime, a constant, and a value, returning the next step in the dynamical system
func dynamicOperator(prime int, constant int, value int) int {
	return (((value * value) + constant) % prime)
}

func writeIt(data []string, writer *csv.Writer) {
	err := writer.Write(data)
	checkError("Write to file failed. ", err)
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

//**********************Old Stuff, refactor or delete***********************************

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

//intSqrt takes an integer, takes the square root, and returns the floor function of the result
func intSqrt(input int) int {
	return int(math.Floor(math.Sqrt(float64(input))))
}
