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
	"sync"
	//"time"
)

type preP [2]int

//critHashEntry is passed to the hash writer function and contains the index as a length 2 array
//and the constant to be written.
type critHashEntry struct {
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

//scraperCounter is a counter that keeps track of scraper time data for preperiodic portraits.
type scraperCounter struct {
	tMax            int
	tSum            int
	singletonSum    int
	nonsingletonSum int
}

//preperiodicOutputData is data ready to be put through a wrapper for writing to a csv.
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
	fmt.Println("starting")
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
	preperiodicChan := make(chan []string)
	go writeIt(preperiodicWriter, preperiodicChan)
	//* Select your primes, here.
	for {
		var nextPrimeWG sync.WaitGroup
		p := <-primeChan
		go computePrimeStats(p, &nextPrimeWG, preperiodicChan)
		nextPrimeWG.Wait()
	}
}

//computePrimeStats takes a prime and builds all the statistics relating to that prime.
//It puts the stats onto various channels that have writers for different output files.
//nextPrimeWG lets the calling function know when to start the next prime computation.
//this should occur after completing critHash writing but before critHash scoring.

func computePrimeStats(prime int, nextPrimeWG *sync.WaitGroup, preperiodicChan chan<- []string) {
	nextPrimeWG.Add(prime - 1)
	critHashEntryChan := make(chan critHashEntry)

	go critHashWriter(prime, nextPrimeWG, critHashEntryChan, preperiodicChan)

	for constant := 1; constant < prime; constant++ {
		go buildPreperiodicPortrait(prime, constant, nextPrimeWG, critHashEntryChan)
	}
	nextPrimeWG.Wait()
	close(critHashEntryChan)
}

//buildPreperiodicPortrait takes a prime and a constant, returning a funcGraph with constant, critHeight, and critCycleLength filled in.
func buildPreperiodicPortrait(prime, constant int, nextPrimeWG *sync.WaitGroup, critHashEntryChan chan<- critHashEntry) {
	critHeight, critCycleLength := critHeightAndCycle(prime, constant)
	critHashEntryChan <- critHashEntry{critHeight, critCycleLength, constant}
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

//critHashWriter is run as a goroutine. It takes a channel of hashEntries, and writes them to the hash as they come in.
//This funtion also initializes its own hash.
func critHashWriter(prime int, nextPrimeWG *sync.WaitGroup, critHashEntryChan <-chan critHashEntry, preperiodicChan chan<- []string) {
	hash := make(map[preP][]int)

	counter := preperiodicCounter{0, 0, 0, 1, 1, 1} //initialie counter after c=0 is accounted for
	for i := 0; i < prime-1; i++ {
		a := <-critHashEntryChan //a is a hash entry from the channel
		incrementPreperiodicCounter(&counter, a)
		hash[[2]int{a.h, a.n}] = append(hash[[2]int{a.h, a.n}], a.constant)
	}
	out := scorecritHash(prime, hash, counter)
	fmt.Println(out.prime)
	preperiodicChan <- []string{strconv.Itoa(out.prime), strconv.FormatFloat(out.hAvg, 'f', -1, 64), strconv.Itoa(out.hMax), strconv.FormatFloat(out.nAvg, 'f', -1, 64), strconv.Itoa(out.nMax), strconv.FormatFloat(out.hnAvg, 'f', -1, 64), strconv.Itoa(out.hnMax), strconv.FormatFloat(out.tAvg, 'f', -1, 64), strconv.Itoa(out.tMax), strconv.FormatFloat(out.singletonRatio, 'f', -1, 64), strconv.Itoa(out.nonsingletonClasses)}

	nextPrimeWG.Wait()
}

func incrementPreperiodicCounter(counter *preperiodicCounter, entry critHashEntry) {

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

func scorecritHash(prime int, hash map[preP][]int, counter preperiodicCounter) preperiodicOutputData {

	//initialize out with prime, hMax, nMax, and hnMax.
	out := preperiodicOutputData{prime, 0, counter.hMax, 0, counter.nMax, 0, counter.hnMax, 0, 0, 0, 0}
	graphs := funcGraphPackage{prime, make([]funcGraph, 0)}
	//need to set hAvg, nAvg, hnAvg, tAvg, tMax, singletonRatio and nonsingletonClasses.
	//set hAvg
	out.hAvg = float64(counter.hSum) / float64(prime)
	out.nAvg = float64(counter.nSum) / float64(prime)
	out.hnAvg = float64(counter.hnSum) / float64(prime)
	out.tAvg, out.tMax, out.singletonRatio, out.nonsingletonClasses, graphs.graphs = hashScraper(hash)
	return out
}

func hashScraper(hash map[preP][]int) (float64, int, float64, int, []funcGraph) {

	//also return []funcGraph
	counter := scraperCounter{0, 0, 0, 0}
	graphSlice := make([]funcGraph, 0)
	for i := range hash {
		if len(hash[i]) == 1 {
			counter.singletonSum++ //increment singletonSum
		} else if len(hash[i]) == 2 {
			counter.nonsingletonSum++ //increment nonsingletonSum
			newT := len(hash[i])
			counter.tSum = counter.tSum + newT //increment tSum
			if counter.tMax < newT {
				counter.tMax = newT //update tMax
			}
			graphSlice = append(graphSlice, funcGraph{hash[i], i[0], i[1]})
		}
	}
	tAvg := float64(counter.tSum) / float64(counter.nonsingletonSum)
	singletonRatio := float64(counter.singletonSum) / float64(counter.singletonSum+counter.nonsingletonSum)

	return tAvg, counter.tMax, singletonRatio, len(graphSlice), graphSlice
}

//dynamicOperator takes a prime, a constant, and a value, returning the next step in the dynamical system
func dynamicOperator(prime int, constant int, value int) int {
	return (((value * value) + constant) % prime)
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

//checkError returns fatal and an error message, given by the string.
func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

//writeIt abstracts error handling for writers away from the programer.
func writeIt(writer *csv.Writer, writerChan <-chan []string) {
	for {
		err := writer.Write(<-writerChan)
		checkError("Write to file failed. ", err)
	}
}
