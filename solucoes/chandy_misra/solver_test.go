package chandy_misra

import (
	"testing"
	"time"

	"github.com/William-SWS/concurrent-programming-work/core"
)

// TestSemDeadlock roda a simulação nos 3 grafos, com tempo
// acelerado, e falha se algum não concluir (possível deadlock) ou se algum
// filósofo não beber o número exigido de vezes. Rode com -race para também
// pegar data races:
//
//	go test -race ./solucoes/chandy_misra/
func TestSemDeadlock(t *testing.T) {
	old := Unidade
	Unidade = 2 * time.Millisecond
	defer func() { Unidade = old }()

	casos := []struct {
		path   string
		n      int
		rounds int
	}{
		{"../../data/caso1_jantar_5.txt", 5, 4},
		{"../../data/caso2_bar_6.txt", 6, 4},
		{"../../data/caso3_bar_12.txt", 12, 3},
	}

	for _, c := range casos {
		t.Run(c.path, func(t *testing.T) {
			g, err := core.LoadGraph(c.path)
			if err != nil {
				t.Fatal(err)
			}

			done := make(chan []*core.Philosopher, 1)
			go func() { done <- New().Run(g, c.rounds) }()

			select {
			case phs := <-done:
				if len(phs) != c.n {
					t.Fatalf("esperava %d filósofos, obtive %d", c.n, len(phs))
				}
				for _, p := range phs {
					if p.Metrics.Drinks != c.rounds {
						t.Errorf("filósofo %d bebeu %d vezes, esperado %d", p.ID, p.Metrics.Drinks, c.rounds)
					}
				}
			case <-time.After(30 * time.Second):
				t.Fatal("timeout: possível deadlock")
			}
		})
	}
}
