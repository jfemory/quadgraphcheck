package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"strings"
)

//funcGraph is a struct that holds the reduced graph information for each constant
//for a given prime. The prime is not included as part of the constPackage, as there
//is an array of these associated with a given prime.
type preP struct {
	constant        int
	critCycleLength int
	critHeight      int
	//components      []block
	//tempList	[][]int //edgeset is temporary and decrements as components are built.
}

func main() {
	p := 11
	portrait := make(chan preP)
	var wg sync.WaitGroup
	for i := 1; i < p; i++ {
		preperiod(p, i, portrait, &wg)
	}
	for i := 1; i < p; i++ {
		fmt.Println(<-portrait)
	}
	wg.Wait()
	close(portrait)
}

//preperiod takes a prime p and a constant c, putting a preP
//onto the portrait chan
func preperiod(p int, c int, portrait chan<- preP, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
	cycleCheck := make([]int, 0)
	cycleCheck = append(cycleCheck, 0)
	var new int
	for i := 0; i < p; i++ {
		new = (cycleCheck[i]*cycleCheck[i] + c) % p
		//fmt.Println(cycleCheck)
		for j := 0; j < len(cycleCheck); j++ {
			if new == cycleCheck[j] {
				portrait <- preP{c, (len(cycleCheck) - j), j}
				return
			}
		}
		cycleCheck = append(cycleCheck, new)
	}
}()
wg.Done()
}

//parsePrimeList takes a list of primes and pushes them one by one onto primeChan
func parsePrimeList(primeChan chan int) {
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
	close(primeChan)
	//	checkError("bufio problem, figure it out...", err)
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
