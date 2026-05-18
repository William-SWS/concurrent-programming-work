package graph

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func ReadMatrix(filePath string) ([][]int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matrix [][]int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		var row []int
		for _, f := range fields {
			val, err := strconv.Atoi(f)
			if err != nil {
				return nil, err
			}
			row = append(row, val)
		}
		matrix = append(matrix, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matrix, nil
}
