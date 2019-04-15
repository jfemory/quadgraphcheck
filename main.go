//Every vertex has out degree 1
//Every vertex has in degree at most 2
//bound of number of cycles of given length
//bound of number of cycles for a given
//Find published version of tree iso paper

//read tree paper, write up a brief review, present in two weeks.
//can I do cannonical encodings on trees that aren't binary trees?
//base d number system to encode d-trees? Think through it.

//Ideas of the paper, prove the theorem. To a uninitiated audience.

/*#quadgraphcheck
Quadratic Functional Graph Isomorphism Checker
The quadratic functional graph isomorphism checker takes a prime number and checks
to see if any of the associated functional graphs are isomorphic. Uploads will be
coming over the next few days.
Place a list of primes in a folder called "list/list.prime" under the
source directory.*/
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"
)

type preP [2]int

//critHashEntry is passed to he hash writer function and contains the index as a length 2 array
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
		computePrimeStats(p, &nextPrimeWG, preperiodicChan)
		fmt.Println(p)
		nextPrimeWG.Wait()
	}
}

//computePrimeStats takes a prime and builds all the statistics relating to that prime.
//It puts the stats onto various channels that have writers for different output files.
//nextPrimeWG lets the calling function know when to start the next prime computation.
//this should occur after completing critHash writing but before critHash scoring.

func computePrimeStats(prime int, nextPrimeWG *sync.WaitGroup, preperiodicChan chan<- []string) {
	nextPrimeWG.Add(prime)
	critHashEntryChan := make(chan critHashEntry)

	go critHashWriter(prime, nextPrimeWG, critHashEntryChan, preperiodicChan)

	for constant := 1; constant < prime; constant++ {
		go buildPreperiodicPortrait(prime, constant, nextPrimeWG, critHashEntryChan)
	}
	nextPrimeWG.Wait()
	close(critHashEntryChan)
}

//dynamicOperator takes a prime, a constant, and a value, returning the next step in the dynamical system
func dynamicOperator(prime int, constant int, value int) int {
	return (((value * value) + constant) % prime)
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
