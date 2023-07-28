package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: go run main.go <input_file>")
		return
	}

	inputFileName := os.Args[1]

	// Initialize nodes from input
	nodes := make([]*Node, 0)
	file, err := os.Open(inputFileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Text()
		tupleData := strings.Split(data[1:len(data)-1], "),(")
		for _, nodeDataStr := range tupleData {
			nodeData := strings.Split(nodeDataStr, ",")
			x, _ := strconv.Atoi(strings.TrimSpace(nodeData[0]))
			y, _ := strconv.Atoi(strings.TrimSpace(nodeData[1]))
			energyLevel, _ := strconv.Atoi(strings.TrimSpace(nodeData[2]))
			nodes = append(nodes, NewNode(x, y, energyLevel))
		}
	}

	// Initial leader selection and group formation
	groups := formGroups(nodes)
	fmt.Println("\nInitial list of nodes:\n")
	for _, group := range groups {
		for _, node := range group {
			fmt.Println(node.toString())
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(nodes))

	// Start the simulation by running each node as a separate goroutine
	for _, group := range groups {
		for _, node := range group {
			go func(node *Node) {
				node.run()
				wg.Done()
			}(node)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("\nSimulation finished!")
}
