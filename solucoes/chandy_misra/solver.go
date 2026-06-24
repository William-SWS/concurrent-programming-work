// Package chandy_misra implementa a Solução 3: Chandy-Misra (garrafas com
// estado, distribuído por troca de mensagens). Cada garrafa está cheia ou
// vazia (inicialmente vazia). Ao ficar com sede, o filósofo envia pedidos aos
// vizinhos; o detentor entrega a garrafa quando não está bebendo; pedidos
// concorrentes são servidos na ordem de chegada.
package chandy_misra

import "github.com/William-SWS/concurrent-programming-work/core"

type Solver struct{}

// New devolve o solver desta solução.
func New() core.Solver { return &Solver{} }

func (s *Solver) Name() string { return "Chandy-Misra" }

func (s *Solver) Run(g *core.Graph, rounds int) []*core.Philosopher {
	// TODO: implementar.
	panic("solucao 'chandy_misra' ainda nao implementada")
}
