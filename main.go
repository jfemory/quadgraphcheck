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

//preP is a struct representing the preperiodic portrait of the critical component of a functional graph.
type preP struct {
	constant        []int
	critHeight      int
	critCycleLength int
	//components      []block
	//tempList	[][]int //edgeset is temporary and decrements as components are built.
}
type outputData struct {
	p                   int
	h_avg               float64
	h_max               int
	n_avg               float64
	n_max               int
	t_avg               float64
	t_max               int
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
	writer.Write([]string{"p", "h_avg", "h_max", "n_avg", "n_max", "t_avg", "t_max", "singleton_ratio", "nonsingletonClasses"})
	primeChan := make(chan int)

	go parsePrimeListCSV(primeChan)

	var waitToScore sync.WaitGroup

	for {
		portChan := make(chan []preP)
		p := <-primeChan
		go buildPrimePortrait(p, portChan, &waitToScore)
		go scorePrimePortrait(p, portChan, writer)
		waitToScore.Wait()
	}
}

//scorePrimePortrait takes a prime portrait slice, port, and puts outputData onto the outData channel
func scorePrimePortrait(p int, portChan <-chan []preP, writer *csv.Writer) {
	port := <-portChan
	var out outputData
	h_max := 0
	h_sum := 0
	n_max := 1
	n_sum := 1
	t_sum := 0
	t_count := 0
	t_max := 0
	singleton_count := 1
	out.p = p

	for i := 0; i < len(port); i++ {
		//fmt.Println("a")
		//set coefficient for sums
		coeff := len(port[i].constant)
		//increment x_sum and x_count
		h_sum = h_sum + (coeff * port[i].critHeight)
		n_sum = n_sum + (coeff * port[i].critCycleLength)
		if len(port[i].constant) != 1 {
			t_sum = t_sum + len(port[i].constant)
			t_count++
		} else {
			singleton_count++
		}
		//update x_max
		if h_max < port[i].critHeight {
			h_max = port[i].critHeight
		}
		if n_max < port[i].critCycleLength {
			n_max = port[i].critCycleLength
		}
		if t_max < len(port[i].constant) {
			t_max = len(port[i].constant)
		}
	}
	out.h_avg = float64(h_sum) / float64(p)
	out.h_max = h_max
	out.n_avg = float64(n_sum) / float64(p)
	out.n_max = n_max
	if t_count == 0 {
		out.t_avg = 1
	} else {
		out.t_avg = float64(t_sum) / float64(t_count)
	}
	out.t_max = t_max
	out.singletonRatio = float64(singleton_count) / float64(p)
	out.nonsingletonClasses = t_count
	fmt.Println(out.p)
	writeIt([]string{strconv.Itoa(out.p), strconv.FormatFloat(out.h_avg, 'f', -1, 64), strconv.Itoa(out.h_max), strconv.FormatFloat(out.n_avg, 'f', -1, 64), strconv.Itoa(out.n_max), strconv.FormatFloat(out.t_avg, 'f', -1, 64), strconv.Itoa(out.t_max), strconv.FormatFloat(out.singletonRatio, 'f', -1, 64), strconv.Itoa(out.nonsingletonClasses)}, writer)
}

//buildPrimePortrait builds an array of preperiodic portraits from channel of preperiodic portraits, as they come in. It returns this array.
func buildPrimePortrait(p int, portChan chan<- []preP, waitToScore *sync.WaitGroup) {
	waitToScore.Add(1)
	primePortrait := make([]preP, 0)
	portrait := make(chan preP)
	var wg sync.WaitGroup
	for i := 1; i < p; i++ {
		go preperiod(p, i, portrait, &wg)
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

//preperiod takes a prime p and a constant c, putting a preP onto the portrait chan. Run as a go routine.
func preperiod(p int, c int, portrait chan<- preP, wg *sync.WaitGroup) {
	wg.Add(1)
	//fmt.Println(c)
	cycleCheck := make([]int, 0)
	cycleCheck = append(cycleCheck, 0)
	var new int
	for i := 0; i < p; i++ {
		new = (cycleCheck[i]*cycleCheck[i] + c) % p
		for j := 0; j < len(cycleCheck); j++ {
			if new == cycleCheck[j] {
				portrait <- preP{[]int{c}, (len(cycleCheck) - j), j}
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
