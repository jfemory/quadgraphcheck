/*
preperiod.go holds all the functions related to generating statistics on a preperiodic portrait.
buildPreperiodicPortrait is the entry point of the preperiod.go file.
*/
package main

import (
	"strconv"
	"sync"
)

//buildPreperiodicPortrait takes a prime and a constant, returning a funcGraph with constant, critHeight, and critCycleLength filled in.
func buildPreperiodicPortrait(prime, constant int, nextPrimeWG *sync.WaitGroup, critHashEntryChan chan<- critHashEntry) {
	critHeight, critCycleLength := critHeightAndCycle(prime, constant)
	critHashEntryChan <- critHashEntry{critHeight, critCycleLength, constant}
	nextPrimeWG.Done()
}

//TODO: buildPreperiodicPortrait should return a slice of nonsingleton equivalence classes of funcGraphs

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
	hash := make(map[preP][]int, prime)

	counter := preperiodicCounter{0, 0, 0, 1, 1, 1} //initialie counter after c=0 is accounted for
	for i := 0; i < prime-1; i++ {
		a := <-critHashEntryChan //a is a hash entry from the channel
		incrementPreperiodicCounter(&counter, a)
		hash[[2]int{a.h, a.n}] = append(hash[[2]int{a.h, a.n}], a.constant)
	}
	out := scorecritHash(prime, hash, counter)
	preperiodicChan <- []string{strconv.Itoa(out.prime), strconv.FormatFloat(out.hAvg, 'f', -1, 64), strconv.Itoa(out.hMax), strconv.FormatFloat(out.nAvg, 'f', -1, 64), strconv.Itoa(out.nMax), strconv.FormatFloat(out.hnAvg, 'f', -1, 64), strconv.Itoa(out.hnMax), strconv.FormatFloat(out.tAvg, 'f', -1, 64), strconv.Itoa(out.tMax), strconv.FormatFloat(out.singletonRatio, 'f', -1, 64), strconv.Itoa(out.nonsingletonClasses)}

	nextPrimeWG.Done()
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
