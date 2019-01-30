package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
	constant        int
	critCycleLength int
	critHeight      int
	components      []block
	//tempList	[][]int //edgeset is temporary and decrements as components are built.
}

type block struct {
	cycleLength int
	trees       []int
}

type outputData struct {
	p          int
	h_avg      float64
	h_max      int
	n_avg      float64
	n_max      int
	t_avg      float64
	t_max      int
	singletons int
}

func main() {
	//new file logic
	file, err := os.Create("preperiodicPortraitStats.csv")
	checkError("Cannot Create File", err)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	data := make(chan []string)
	primeChan := make(chan int)
	listEmpty := make(chan bool)

	go parsePrimeList(primeChan, listEmpty)
	go generatePreperiodStats(data, primeChan)
	go writeIt(data, writer) //write strings to file

	//Primes := []int{3, 5, 7}
	//Primes := []int{17}
	//for i := 0; i < len(Primes); i++ {
	//	pack, err := initialSort(Primes[i])
	//	checkError("Prime Sort Failed.", err)
	//	fmt.Println(pack)
	//}
	fmt.Println("Preparing Output")
	<-listEmpty
	fmt.Println("Done")
}

func generatePreperiodStats(data chan []string, primeChan chan int) {
	data <- []string{"p", "h_avg", "h_max", "n_avg", "n_max", "t_avg", "t_max", "total_singleton"}
	var output outputData
	for {
		prime := <-primeChan
		pack, err := initialSort(prime)
		//fmt.Println(pack)
		checkError("Problem sorting prime.", err)
		output.p = prime
		getH(&pack, &output)
		getN(&pack, &output)
		getT(&pack, &output)
		data <- []string{strconv.Itoa(output.p), strconv.FormatFloat(output.h_avg, 'f', -1, 64), strconv.Itoa(output.h_max), strconv.FormatFloat(output.n_avg, 'f', -1, 64), strconv.Itoa(output.n_max), strconv.FormatFloat(output.t_avg, 'f', -1, 64), strconv.Itoa(output.t_max), strconv.Itoa(output.singletons)}
	}
}

func writeIt(data chan []string, writer *csv.Writer) {
	for {
		output := <-data
		err := writer.Write(output)
		checkError("Write to file failed.", err)
	}
}

func getH(pack *primePackage, output *outputData) error {
	var h_max int
	var h_sum int
	h_max = 0
	h_sum = 0
	for i := 0; i < len(pack.graphs); i++ {
		if h_max < pack.graphs[i][0].critHeight {
			h_max = pack.graphs[i][0].critHeight
		}
		h_sum = h_sum + (len(pack.graphs[i]) * pack.graphs[i][0].critHeight)
	}
	output.h_avg = float64(h_sum) / float64(pack.prime)
	output.h_max = h_max
	return nil
}

func getN(pack *primePackage, output *outputData) error {
	var n_max int
	var n_sum int
	n_max = 0
	n_sum = 0
	for i := 0; i < len(pack.graphs); i++ {
		if n_max < pack.graphs[i][0].critCycleLength {
			n_max = pack.graphs[i][0].critCycleLength
		}
		n_sum = n_sum + (len(pack.graphs[i]) * pack.graphs[i][0].critCycleLength)
	}
	output.n_avg = float64(n_sum) / float64(pack.prime)
	output.n_max = n_max
	return nil
}

//T is for tuple, size of equivalence classes
func getT(pack *primePackage, output *outputData) error {
	var t_max int
	var t_sum int
	var singletonCount int
	t_max = 1
	t_sum = 1
	singletonCount = 1
	for i := 0; i < len(pack.graphs); i++ {
		if t_max < len(pack.graphs[i]) {
			t_max = len(pack.graphs[i])
		}
		if len(pack.graphs[i]) == 1 {
			singletonCount++
		}
		t_sum = t_sum + (len(pack.graphs[i]))
	}
	output.t_avg = float64(t_sum) / float64(len(pack.graphs))
	output.t_max = t_max
	output.singletons = singletonCount
	return nil
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
	return primePackage{int(p), sortedGraphs}, err
}

//buildFuncGraph takes a prime, p, and constant constant, and returns a prime package
//populated with cycle length
func buildFuncGraph(p int, constant int) (funcGraph, error) {
	critCycleLength, critHeight, err := easyCycleCheck(p, constant)
	return funcGraph{constant, critCycleLength, critHeight, nil}, err
}

//easyCycleCheck takes a prime, p, and a constant, constant. It returns
//the cycle length and critical point height, in that order.
//An error is also returned if out of bounds.
func easyCycleCheck(p int, constant int) (int, int, error) {
	cycleCheck := make([]int, 0)
	cycleCheck = append(cycleCheck, 0)
	var new int
	for i := 0; i < p; i++ {
		new = (cycleCheck[i]*cycleCheck[i] + constant) % p
		//fmt.Println(cycleCheck)
		for j := 0; j < len(cycleCheck); j++ {
			if new == cycleCheck[j] {
				return len(cycleCheck) - j, j, nil
			}
		}
		cycleCheck = append(cycleCheck, new)
	}
	return -1, -1, errors.New("easyCycleCheck: Index out of bounds.")
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

//parsePrimeList takes a list of primes and pushes them one by one onto primeChan
func parsePrimeList(primeChan chan int, listEmpty chan bool) {
	//open file logic
	openFile, err := os.Open("list.prime")
	checkError("Failed to open prime list file.", err)
	defer openFile.Close()

	scanner := bufio.NewScanner(openFile)
	for scanner.Scan() {
		stringSlice := (strings.Split(strings.Trim(scanner.Text(), "[]"), " "))
		for i := 0; i < len(stringSlice); i++ {
			prime, err := strconv.Atoi(stringSlice[i])
			checkError("Problem converting file into primes.", err)
			fmt.Println(prime)
			primeChan <- prime
		}
	}
	listEmpty <- true
	//	checkError("bufio problem, figure it out...", err)
}
