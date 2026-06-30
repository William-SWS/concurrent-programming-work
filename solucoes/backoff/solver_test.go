package backoff

import (
	"testing"

	"github.com/William-SWS/concurrent-programming-work/core"
)

// TestSemDeadlock roda a simulação num grafo pequeno e falha se houver
// deadlock/timeout. Rode com -race para também pegar data races:

func TestSemDeadlock(t *testing.T) {

	g, err := core.LoadGraph("../../data/caso1_jantar_5.txt")
	if err != nil {
		t.Fatal(err)
	}
	
	// Chama a função New() do seu solver.go e roda a simulação
	phs := New().Run(g, 2)
	
	// Verifica se a quantidade de filósofos processados bate com o Grafo
	if len(phs) != g.N {
		t.Fatalf("esperava %d filosofos, obtive %d", g.N, len(phs))
	}
}
