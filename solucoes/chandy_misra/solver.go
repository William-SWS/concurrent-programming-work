// Package chandy_misra implementa a Solução 3: Chandy-Misra, a solução do
// "drinking philosophers problem" por troca de mensagens (Seção 5.2 do artigo
// de Chandy e Misra, 1984).
//
// São duas camadas:
//
//   - Garrafas (recurso real): o filósofo com sede pede aos vizinhos as
//     garrafas do subconjunto que sorteou; ao receber um pedido cede a garrafa,
//     a menos que precise dela e tenha precedência (ou esteja bebendo). Pedidos
//     são servidos na ordem de chegada (request token + mailbox).
//
//   - Garfos (auxiliares): implementam o grafo de precedência H via o "jantar
//     dos filósofos higiênico" (garfo limpo/sujo). Para ter precedência o
//     filósofo precisa segurar todos os garfos incidentes (estado "comendo"),
//     e é isso que mantém H acíclico. Sem essa camada, beber de um subconjunto
//     pode tornar H cíclico e causar deadlock.
//
// A distribuição inicial (garfo e garrafa sujos/no filósofo de menor id) torna
// H acíclico, garantindo ausência de deadlock e de starvation. Comunicação é só
// por mensagens (channels); não há estado compartilhado mutável entre
// goroutines (validado com `go test -race`).
package chandy_misra

import (
	"sync"
	"time"

	"github.com/William-SWS/concurrent-programming-work/core"
)

// Unidade é a escala de tempo da simulação. Por padrão 1 segundo, conforme o
// enunciado (tranquilo: 0..n s; bebendo: 1 s). Os testes reduzem esse valor
// pra rodar rápido.
var Unidade = time.Second

type Solver struct{}

// New devolve o solver dessa solução.
func New() core.Solver { return &Solver{} }

func (s *Solver) Name() string { return "Chandy-Misra" }

func (s *Solver) Run(g *core.Graph, rounds int) []*core.Philosopher {
	sim := &simulacao{
		rounds: rounds,
		stop:   make(chan struct{}),
		inbox:  make([]chan message, g.N),
	}
	for i := range sim.inbox {
		sim.inbox[i] = make(chan message, 4*g.N+8)
	}

	phils := make([]*phil, g.N)
	out := make([]*core.Philosopher, g.N)
	for i := 0; i < g.N; i++ {
		phils[i] = newPhil(i, g, sim)
		out[i] = phils[i].core
	}

	// Distribuição inicial acíclica: pra cada aresta, o filósofo de menor id
	// começa com o garfo (sujo) e com a garrafa; o de maior id fica com os dois
	// request tokens. Garfos sujos sempre no menor id => grafo de precedência H
	// acíclico => sem deadlock.
	for i := 0; i < g.N; i++ {
		for _, j := range g.Neighbors(i) {
			e := phils[i].edges[j]
			if i < j {
				e.fork = true
				e.dirty = true
				e.bot = true
			} else {
				e.reqf = true
				e.reqb = true
			}
		}
	}

	sim.completed.Add(g.N)
	sim.exited.Add(g.N)
	for _, p := range phils {
		go p.run()
	}

	sim.completed.Wait() // todos beberam `rounds` vezes
	close(sim.stop)      // libera os filosofos "aposentados" que ainda serviam
	sim.exited.Wait()

	return out
}

// simulacao guarda o estado global imutável durante a execução (channels e
// grupos de espera). Nenhum campo aqui é mutado por mais de uma goroutine.
type simulacao struct {
	rounds    int
	stop      chan struct{}
	inbox     []chan message
	completed sync.WaitGroup // conta filósofos que cumpriram suas rodadas
	exited    sync.WaitGroup // conta goroutines que já retornaram
}
