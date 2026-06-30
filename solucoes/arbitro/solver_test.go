package arbitro

import (
	"testing"

	"github.com/William-SWS/concurrent-programming-work/core"
)

// TestSemDeadlock roda a simulação num grafo pequeno e falha se houver
// deadlock/timeout. Rode com -race para também pegar data races:
//
//	go test -race ./solucoes/arbitro/
func TestSemDeadlock(t *testing.T) {
	t.Skip("TODO: implementar Solver.Run e remover este Skip")

	g, err := core.LoadGraph("../../data/caso1_jantar_5.txt")
	if err != nil {
		t.Fatal(err)
	}
	phs := New().Run(g, 2)
	if len(phs) != g.N {
		t.Fatalf("esperava %d filosofos, obtive %d", g.N, len(phs))
	}
}
