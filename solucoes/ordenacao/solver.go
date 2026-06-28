// Package ordenacao implementa a Solução 1: Numeração das Garrafas
// (ordenação de recursos). Cada garrafa (aresta) recebe um número e o filósofo
// sempre adquire primeiro a garrafa de menor número, evitando espera circular
// e, portanto, deadlock.
package ordenacao

import "github.com/William-SWS/concurrent-programming-work/core"

type Solver struct{}

// New devolve o solver desta solução.
func New() core.Solver { return &Solver{} }

func (s *Solver) Name() string { return "Ordenação de Recursos" }

func (s *Solver) Run(g *core.Graph, rounds int) []*core.Philosopher {
	// TODO: implementar.
	panic("solucao 'ordenacao' ainda nao implementada")
}
