package mail 

//turn a binary tree into a binary number
//left weight them to make unique
//XOR the results for an isomorphism check
import (
	"fmt"
	"log"
	"os"
)

//PrimePackage is the wrapper for all the graphs of a given prime
type PrimePackage struct {
	Prime     int
	GraphList []Graph
}

//A Graph is a collection of vertices and edges connecting those vertices
type Graph struct {
	C      int
	Edges  []int
	Blocks []Block
}

//A Block is a subgraph of a grpah, disjoint, a component size,
//a cycle length, and aa flag for whether the component contains 0.
type Block struct {
	Size         int
	CycleLength  int
	ZeroDistance int
	BlockEdges   [][]int
}

func main() {

	file, err := os.Create("output-GraphFinder0.2.txt")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	//End file logic
	GenerateGraphSet(4001)
	fmt.Fprintln(file, "test")
}

//GenerateGraphSet builds a slice of graphs from c=1 to c=p
//for a prime p
func GenerateGraphSet(prime int) PrimePackage {
	graphList := InitializeGraphList(prime)
	//Now, we need to block sort each graph. Loop to handle iteration over c,
	//using a tmplist for each c to facilitate block sorting
	for i := 0; i < len(graphList); i++ {
		sortMe := &graphList[i]
		//break this out into new funciton
		tempList := MakeTempList(sortMe.Edges)
		graphList[i].Blocks = BlockListBuilder(tempList)
	}
	var packegedOutput PrimePackage
	packegedOutput.GraphList = graphList
	packegedOutput.Prime = prime
	for i := 0; i < len(packegedOutput.GraphList); i++ {
		fmt.Println(packegedOutput.GraphList[i])
	}
	//fmt.Println(packegedOutput)
	return packegedOutput
}

//BlockListBuilder builds all of the blocks of a graph of given prime p
func BlockListBuilder(tempList [][]int) []Block {
	var blockToAppend Block
	var blockList []Block
	for {
		if len(tempList) == 0 {
			return blockList
		}

		tempList, blockToAppend = NewBlock(tempList)
		blockList = append(blockList, blockToAppend)
		//	fmt.Println(tempList)
		//	fmt.Println(blockList)

	}
}

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

//idea for setting zero flag that isn't working. Needs to be set in stepDown because we have
//size of block starting at zero without stepup.
//if blockEdges[0][0] == 0 {
//	outputBlock.ZeroDistance = len(outputBlock.BlockEdges) - outputBlock.CycleLength
//} else {
//	outputBlock.ZeroDistance = -1
//}

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

//MakeTempList creates a tempList from an edgeset
func MakeTempList(edgeSet []int) [][]int {
	length := len(edgeSet)
	tempList := make([][]int, length)
	for i := 0; i < length; i++ {
		tempList[i] = make([]int, 2)
		tempList[i][0] = i
		tempList[i][1] = edgeSet[i]
	}
	return tempList
}

///////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////
////////////////////Initialization Moving Parts////////////////////
///////////////////////////////////////////////////////////////////

//InitializeGraphList makes a new graph list for a prime p
func InitializeGraphList(prime int) []Graph {
	graphList := make([]Graph, prime-1)
	for i := 1; i < prime; i++ {
		graphList[i-1] = InitializeGraph(prime, i)
	}
	return graphList
}

//InitializeGraph creates a new graph of prime p and constant c
func InitializeGraph(prime int, c int) Graph {
	var graph Graph
	graph.C = c
	graph.Edges = make([]int, prime)
	for i := 0; i < prime; i++ {
		graph.Edges[i] = (i*i + c) % prime
	}

	return graph
}
