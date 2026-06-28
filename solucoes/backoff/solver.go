// Package backoff implementa a Solução 4: Aleatoriedade (Randomized Backoff).
// Se o filósofo não consegue todas as garrafas, ele solta as que pegou e espera
// um tempo aleatório antes de tentar de novo, quebrando a espera circular de
// forma probabilística.
package backoff

import "github.com/William-SWS/concurrent-programming-work/core"

type Solver struct{}

// New devolve o solver desta solução.
func New() core.Solver { return &Solver{} }

func (s *Solver) Name() string { return "Randomized Backoff" }

func (s *Solver) Run(g *core.Graph, rounds int) []*core.Philosopher {
	// TODO: implementar.
	panic("solucao 'backoff' ainda nao implementada")
}
