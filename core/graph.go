package core

import (
	"fmt"
	"os"
	"strings"
)

// Graph é o grafo do bar (vértices = filósofos, arestas = garrafas),
// representado por uma matriz de adjacência quadrada de ordem N.
type Graph struct {
	N   int
	Adj [][]bool
}

// LoadGraph lê um arquivo .txt onde cada linha é uma linha da matriz de
// adjacência. Qualquer caractere que não seja '0' ou '1' (vírgulas, espaços)
// é ignorado, então tanto "0,1,0" quanto "010" são aceitos.
func LoadGraph(path string) (*Graph, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lendo grafo %q: %w", path, err)
	}

	var rows [][]bool
	for _, line := range strings.Split(string(raw), "\n") {
		var row []bool
		for _, c := range line {
			switch c {
			case '0':
				row = append(row, false)
			case '1':
				row = append(row, true)
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	n := len(rows)
	if n == 0 {
		return nil, fmt.Errorf("grafo %q está vazio", path)
	}
	for i, row := range rows {
		if len(row) != n {
			return nil, fmt.Errorf("matriz não é quadrada: linha %d tem %d colunas, esperado %d", i, len(row), n)
		}
	}
	return &Graph{N: n, Adj: rows}, nil
}

// Neighbors devolve os vértices adjacentes a v (vizinhos com quem v
// compartilha uma garrafa).
func (g *Graph) Neighbors(v int) []int {
	var ns []int
	for j := 0; j < g.N; j++ {
		if g.Adj[v][j] {
			ns = append(ns, j)
		}
	}
	return ns
}

// Degree é o número de arestas adjacentes ao vértice v (o "n" do enunciado).
func (g *Graph) Degree(v int) int {
	return len(g.Neighbors(v))
}

// Edges devolve todas as arestas (garrafas) como pares ordenados {i,j} com i<j.
// Útil para soluções que precisam numerar/identificar cada garrafa.
func (g *Graph) Edges() [][2]int {
	var es [][2]int
	for i := 0; i < g.N; i++ {
		for j := i + 1; j < g.N; j++ {
			if g.Adj[i][j] {
				es = append(es, [2]int{i, j})
			}
		}
	}
	return es
}
