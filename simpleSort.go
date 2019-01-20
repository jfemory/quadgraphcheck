package main

import (
	"fmt"
	//	"log"
	//	"os"
	"errors"
)

//primePackage is a struct that holds a prime and an array of funcGraph
type primePackage struct {
	prime  int
	graphs [][]funcGraph
}

//funcGraph is a struct that holds the reduced graph information for each constant
//for a given prime. The prime is not included as part of the constPackage, as there
//is an array of these associated with a given prime.
type funcGraph struct {
	c           int
	critCycleLength int
	critHeight  int
	trees       []int
}

func main() {
	Primes := []int{3, 5, 7, 11, 13, 17, 19, 23, 29, 32, 37, 41, 43, 47}
	for i := 0; i < len(Primes); i++ {
		output, _ := initialSort(Primes[i])
		fmt.Println(Primes[i])
		for j := 0; j < len(output.graphs); j++ {
			fmt.Println(output.graphs[j])
		}
	}
	//	file, err := os.Create("output-GraphFinder0.2.txt")
	//	if err != nil {
	//		log.Fatal("Cannot create file", err)
	//	}
	//	defer file.Close()
	//End file logic
	//	fmt.Fprintln(file, "test")
}

//initialSort takes a prime, p, and returns a primePackage with an inital sort of
//graphs based on cycle length of the critical component and the critical height.
func initialSort(p int) (primePackage, error) {
	var sortedGraphs [][]funcGraph
	var newGraph funcGraph
	var err error
	for i := 1; i < p; i++ {
		newGraph, err = buildFuncGraph(p, i)
		for j := 0; j < len(sortedGraphs); j++ {
			if sortedGraphs[j][0].critCycleLength == newGraph.critCycleLength && sortedGraphs[j][0].critHeight == newGraph.critHeight {
				sortedGraphs[j] = append(sortedGraphs[j], newGraph)
				i++
			}
		}
		sortedGraphs = append(sortedGraphs, []funcGraph{newGraph})
	}
	return primePackage{p, sortedGraphs}, err
}

//buildFuncGraph takes a prime, p, and constant c, and returns a prime package
//populated with cycle length
func buildFuncGraph(p int, c int) (funcGraph, error) {
	critCycleLength, critHeight, err := cycleCheck(p, c)
	return funcGraph{c, critCycleLength, critHeight, nil}, err
}
 
//cycleCheck takes a prime, p, and a constant, c. It returns
//the cycle length and critical point height, in that order.
//An error is also returned if out of bounds.
func cycleCheck(p int, c int) (int, int, error) {
	cycleCheck := make([]int, 0)
	cycleCheck = append(cycleCheck, 0)
	var new int
	for i := 0; i < p; i++ {
		new = (cycleCheck[i]*cycleCheck[i] + c) % p
		//fmt.Println(cycleCheck)
		for j := 0; j < len(cycleCheck); j++ {
			if new == cycleCheck[j] {
				return len(cycleCheck) - j, j, nil
			}
		}
		cycleCheck = append(cycleCheck, new)
	}
	return -1, -1, errors.New("cycleCheck: Index out of bounds.")
}
