package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/William-SWS/concurrent-programming-work/internal/solution1_ordering"
	"github.com/William-SWS/concurrent-programming-work/pkg/graph"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/main.go <matrix_file_path>")
		return
	}

	filePath := os.Args[1]
	matrix, err := graph.ReadMatrix(filePath)
	if err != nil {
		fmt.Printf("Error reading matrix: %v\n", err)
		return
	}

	// Extract Case ID from filename (e.g., caso1_jantar_5.txt -> 1)
	base := filepath.Base(filePath)
	caseIDStr := ""
	for _, char := range base {
		if char >= '0' && char <= '9' {
			caseIDStr += string(char)
		} else if caseIDStr != "" {
			break
		}
	}
	
	caseID, _ := strconv.Atoi(caseIDStr)
	rounds := 6
	if caseID == 3 {
		rounds = 3
	}

	fmt.Printf("Starting Simulation: Case %d, Rounds %d, Matrix File: %s\n", caseID, rounds, filePath)

	g := graph.New(matrix)

	solver := solution1_ordering.Solver{
		Graph:  g,
		Rounds: rounds,
		CaseID: caseID,
	}

	solver.Run()
}
