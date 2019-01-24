package main

import (
	"fmt"
	//	"log"
		"os"
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
	components  []block
	tempList	[][]int //edgeset is temporary and decrements as components are built. 
}

type block struct {
	cycleLength int
	trees []int
}

type tree struct {
	left *tree
	right *tree
	down *tree
	height int
	weight int
}

//Might need to get rid of this struct
//A Block is a subgraph of a grpah, disjoint, a component size,
//a cycle length, and aa flag for whether the component contains 0.
type Block struct {
	Size         int
	CycleLength  int
	ZeroDistance int
	BlockEdges   [][]int
}

func main() {
	pack, err := initialSort(17)
	if err != nil {
		os.Exit(-1)
	}
	for i := 0; i < len(pack.graphs); i++{
		fmt.Println(pack.graphs[i])
	}
	fmt.Println(" ")
	err = zeroTree(&pack)
	if err != nil {
		os.Exit(-1)
	}
	//for i := 0; i < len(pack.graphs); i++{
//		fmt.Println(pack.graphs[i])
//	}
	//	file, err := os.Create("output-GraphFinder0.2.txt")
	//	if err != nil {
	//		log.Fatal("Cannot create file", err)
	//	}
	//	defer file.Close()
	//End file logic
	//	fmt.Fprintln(file, "test")
}

///////////////////////////////////////////////////////////////////
//////////////////// Tree Isomorphism Check ///////////////////////
///////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////
////////////////////// Zero Tree Iso Check ////////////////////////
///////////////////////////////////////////////////////////////////

//zeroTreeCheck takes a pointer to a prime package and does a zero component
//isomorphism check on any non singleton equivalence class
func zeroTree(inputPackage *primePackage) error{
	for i:= 0; i < len(inputPackage.graphs); i++ {
		if len(inputPackage.graphs[i]) > 1 {
			for j := 0; j < len(inputPackage.graphs[i]); j++{
				err := populateTempList(&inputPackage.graphs[i][j], inputPackage.prime)
				var block Block
				inputPackage.graphs[i][j].tempList, block = NewBlock(inputPackage.graphs[i][j].tempList)
				fmt.Println(block)
				if err != nil {
					return err
				}
			}
//////////////////////End Populate TempList////////////////	
		}
	}
var err error
	return err
}

//populateTempList takes a functional graph pointer and modifies it, returning an error if unsucessful
func populateTempList(graph *funcGraph, p int) error {
	for i := 0; i < p; i++{
		temp0 := i
		temp1 := ((i * i) + graph.c ) % p
		temp := []int{temp0, temp1}
		graph.tempList = append(graph.tempList, temp)
	}
	var err error
	return err
}

///////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////

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

//buildFuncGraph takes a prime, p, and constant c, and returns a prime package
//populated with cycle length
func buildFuncGraph(p int, c int) (funcGraph, error) {
	critCycleLength, critHeight, err := easyCycleCheck(p, c)
	return funcGraph{c, critCycleLength, critHeight, nil, nil}, err
}
 
//easyCycleCheck takes a prime, p, and a constant, c. It returns
//the cycle length and critical point height, in that order.
//An error is also returned if out of bounds.
func easyCycleCheck(p int, c int) (int, int, error) {
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
	return -1, -1, errors.New("easyCycleCheck: Index out of bounds.")
}

///////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////

//NewBlock takes a tempList and constructs a block of a graph
func NewBlock(tempList [][]int) ([][]int, Block) {
	var outputBlock Block
	var tempEdges [][]int
	tempList, outputBlock = stepDown(tempList)
	tempList, tempEdges = stepUp(tempList, outputBlock.BlockEdges)
	//fmt.Println(tempEdges)
	outputBlock.BlockEdges = tempEdges
	outputBlock.Size = len(outputBlock.BlockEdges)
	//outputBlock.BlockEdges = blockEdges
	return tempList, outputBlock
}

//stepUp walks back up the list, finding vertices
func stepUp(tempList [][]int, blockEdges [][]int) ([][]int, [][]int) {
	listLength := len(blockEdges)
	for i := 0; i < listLength; i++ {
		vertex := blockEdges[i][0]
		for k := 0; k < 2; k++ {
			index := FindVertexIndex(vertex, tempList, 1)
			if index == -1 {
				break
			}
			tempList, blockEdges = moveEdgeByIndex(tempList, blockEdges, index)
			//fmt.Println(blockEdges)
			listLength++
		}
	}
	//fmt.Println(blockEdges)
	return tempList, blockEdges
}


//stepDown does teh same thing, hopefully correctly
func stepDown(tempList [][]int) ([][]int, Block) {
	var blockEdges [][]int
	var outputBlock Block
	bigLength := len(tempList)
	//move 0th edge
	tempList, blockEdges = moveEdgeByIndex(tempList, blockEdges, 0)
	cycleFlag := CycleCheck(blockEdges)
	if cycleFlag != 0 {
		outputBlock.BlockEdges = blockEdges
		outputBlock.CycleLength = cycleFlag
		return tempList, outputBlock
	}
	for i := 0; i < bigLength-1; i++ {
		index := findNextIndex(blockEdges[len(blockEdges)-1][1], tempList)
		tempList, blockEdges = moveEdgeByIndex(tempList, blockEdges, index)
		cycleFlag := CycleCheck(blockEdges)
		if cycleFlag != 0 {
			outputBlock.BlockEdges = blockEdges
			outputBlock.CycleLength = cycleFlag
			return tempList, outputBlock
		}
	}
	outputBlock.BlockEdges = blockEdges
	return tempList, outputBlock
}

//findNextIndex
func findNextIndex(vertex int, tempList [][]int) int {
	for i := 0; i < len(tempList); i++ {
		if tempList[i][0] == vertex {
			return i
		}
	}
	return -1
}

//FindVertexIndex takes a vertex and finds the index of the vertex in the
//0 row of the tempList
func FindVertexIndex(vertex int, tempList [][]int, row int) int {
	for i := 0; i < len(tempList); i++ {
		if vertex == tempList[i][row] {
			return i
		}
	}
	return -1
}

//CycleCheck takes a vertex and a block edgeset and determines wheteher we're
//in a cycle
func CycleCheck(blockEdges [][]int) int {
	vertex := blockEdges[len(blockEdges)-1][1]
	for j := 0; j < len(blockEdges); j++ {
		if vertex == blockEdges[j][0] {
			cycleFlag := (len(blockEdges) - j)
			return cycleFlag
		}
	}
	return 0
}

//moveEdgeByIndex moves an edge from tmpList to blockEdges by index in tempList
func moveEdgeByIndex(tempList [][]int, blockEdges [][]int, index int) ([][]int, [][]int) {
	blockEdges = append(blockEdges, tempList[index])
	tempList = RemoveEdgeByIndex(index, tempList)
	return tempList, blockEdges
}

//RemoveEdgeByIndex does just that
func RemoveEdgeByIndex(index int, list [][]int) [][]int {
	list[index] = list[len(list)-1]
	list[len(list)-1] = []int{0, 0}
	list = list[:len(list)-1]
	return list
}