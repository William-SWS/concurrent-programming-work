package ordenacao

import (
	"testing"
	"time"

	"github.com/William-SWS/concurrent-programming-work/core"
)

func TestSemDeadlock(t *testing.T) {
	g, err := core.LoadGraph("../../data/caso1_jantar_5.txt")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan []*core.Philosopher, 1)
	go func() {
		done <- New().Run(g, 1)
	}()

	select {
	case phs := <-done:
		if len(phs) != g.N {
			t.Fatalf("esperava %d filosofos, obtive %d", g.N, len(phs))
		}
		for _, p := range phs {
			if p.Metrics.Drinks != 1 {
				t.Fatalf("filosofo %d bebeu %d vezes, esperado 1", p.ID, p.Metrics.Drinks)
			}
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout: possivel deadlock na solucao por ordenacao")
	}
}
