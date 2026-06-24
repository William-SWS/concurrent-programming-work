// Package arbitro implementa a Solução 2: Árbitro (Garçom). Um árbitro central
// controla quem pode pegar garrafas; o filósofo pede permissão antes de
// adquirir qualquer recurso. O garçom garante no máximo N-1 filósofos
// tentando beber ao mesmo tempo, evitando deadlock.
package arbitro

import "github.com/William-SWS/concurrent-programming-work/core"

type Solver struct{}

// New devolve o solver desta solução.
func New() core.Solver { return &Solver{} }

func (s *Solver) Name() string { return "Árbitro (Garçom)" }

func (s *Solver) Run(g *core.Graph, rounds int) []*core.Philosopher {
	// TODO: implementar.
	panic("solucao 'arbitro' ainda nao implementada")
}
