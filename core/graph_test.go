package core

import "testing"

// TestDataFiles garante que os 3 grafos carregam, têm a ordem
// esperada e são simétricos.
func TestDataFiles(t *testing.T) {
	casos := []struct {
		path string
		n    int
	}{
		{"../data/caso1_jantar_5.txt", 5},
		{"../data/caso2_bar_6.txt", 6},
		{"../data/caso3_bar_12.txt", 12},
	}
	for _, c := range casos {
		g, err := LoadGraph(c.path)
		if err != nil {
			t.Fatalf("%s: %v", c.path, err)
		}
		if g.N != c.n {
			t.Errorf("%s: N=%d, esperado %d", c.path, g.N, c.n)
		}
		for i := 0; i < g.N; i++ {
			if g.Adj[i][i] {
				t.Errorf("%s: vértice %d tem laço para si mesmo", c.path, i)
			}
			for j := 0; j < g.N; j++ {
				if g.Adj[i][j] != g.Adj[j][i] {
					t.Errorf("%s: matriz não simétrica em (%d,%d)", c.path, i, j)
				}
			}
		}
	}
}
