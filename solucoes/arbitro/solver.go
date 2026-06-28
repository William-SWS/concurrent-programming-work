// Package arbitro implementa a Solução 2: Árbitro (Garçom). Um árbitro central
// controla quem pode pegar garrafas; o filósofo pede permissão antes de
// adquirir qualquer recurso. O garçom garante no máximo N-1 filósofos
// tentando beber ao mesmo tempo, evitando deadlock.
//
// FUNCIONAMENTO:
//
//	┌──────────────────────────────────────────────────────────┐
//	│  Árbitro (Garçom)                                        │
//	│  ────────────────                                        │
//	│  Mantém um semáforo de capacidade N-1.                  │
//	│  Philosopher i só pode iniciar a tentativa de adquirir   │
//	│  garrafas se houver vaga no semáforo.                    │
//	│  Quando termina de beber, libera a vaga.                 │
//	└──────────────────────────────────────────────────────────┘
//
// PROVA DE QUE EVITA DEADLOCK:
//
//	Com N filósofos e no máximo N-1 ativos ao mesmo tempo,
//	pelo menos UM filósofo está fora da disputa (tranquilo).
//	Mesmo que todos os N-1 ativos segurem 1 garrafa cada,
//	ainda sobra pelo menos 1 garrafa livre — e o filósofo
//	que está tranquilo não está segurando nada, então ele
//	pode avançar, quebrando qualquer espera circular.
//
// ESTRUTURA:
//   - arbiter: struct com channel de tamanho N-1 (semáforo)
//   - bottle:  sync.Mutex por aresta (garrafa)
//   - Cada filósofo roda em sua própria goroutine
//
// FLUXO POR RODADA (cada filósofo):
//
//  1. TRANQUILO → dorme 0..n segundos (aleatório)
//  2. COM_SEDE  → envia token ao árbitro (bloqueia se N-1 ativos)
//     escolhe aleatoriamente 2..n garrafas para pedir
//     lock() em cada garrafa escolhida (adquire recursos)
//  3. BEBENDO   → dorme 1 segundo (bebendo)
//     unlock() em cada garrafa
//     recebe token do árbitro (libera vaga)
//  4. volta ao passo 1, até completar 'rounds' bebidas
package arbitro

import (
	"math/rand"
	"sync"
	"time"

	"github.com/William-SWS/concurrent-programming-work/core"
)

// ─── Árbitro ────────────────────────────────────────────────────────────────

// arbiter central: garante no máximo N-1 filósofos simultâneos.
type arbiter struct {
	sem chan struct{} // semáforo de capacidade N-1
}

func newArbiter(maxActive int) *arbiter {
	return &arbiter{sem: make(chan struct{}, maxActive)}
}

// acquire bloqueia o filósofo até ter vaga no semáforo.
func (a *arbiter) acquire() {
	a.sem <- struct{}{}
}

// release libera a vaga para outro filósofo.
func (a *arbiter) release() {
	<-a.sem
}

// ─── Garrafa ────────────────────────────────────────────────────────────────

// bottleMap constrói um mutex para cada aresta do grafo.
// Usamos o par ordenado {i,j} com i<j como chave.
func bottleMap(g *core.Graph) map[[2]int]*sync.Mutex {
	bm := make(map[[2]int]*sync.Mutex)
	for i := 0; i < g.N; i++ {
		for j := i + 1; j < g.N; j++ {
			if g.Adj[i][j] {
				bm[[2]int{i, j}] = &sync.Mutex{}
			}
		}
	}
	return bm
}

// ─── Solver ─────────────────────────────────────────────────────────────────

type Solver struct{}

func New() core.Solver { return &Solver{} }

func (s *Solver) Name() string { return "Árbitro (Garçom)" }

func (s *Solver) Run(g *core.Graph, rounds int) []*core.Philosopher {
	philosophers := make([]*core.Philosopher, g.N)
	for i := 0; i < g.N; i++ {
		philosophers[i] = core.NewPhilosopher(i, g)
	}

	bottles := bottleMap(g)
	arb := newArbiter(g.N - 1) // max N-1 ativos
	var wg sync.WaitGroup

	for id := 0; id < g.N; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p := philosophers[id]
			neighbors := g.Neighbors(id)
			n := len(neighbors) // degree

			for drink := 0; drink < rounds; drink++ {

				// ── 1. TRANQUILO ──
				thinkingTime := time.Duration(rand.Intn(n+1)) * time.Second
				thinkStart := time.Now()
				time.Sleep(thinkingTime)
				p.Record(core.Tranquilo, time.Since(thinkStart))

				// ── 2. COM SEDE (tentar adquirir garrafas) ──
				thirstyStart := time.Now()
				arb.acquire() // bloqueia se já há N-1 ativos

				// --- escolher aleatoriamente de 2 a n garrafas ---
				k := 2 + rand.Intn(n-1) // 2..n
				// shuffle e pega os primeiros k vizinhos
				shuffled := make([]int, n)
				copy(shuffled, neighbors)
				rand.Shuffle(n, func(i, j int) {
					shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
				})
				chosen := shuffled[:k]

				// --- adquirir as garrafas escolhidas ---
				// (como o árbitro garante N-1, não há deadlock)
				for _, nb := range chosen {
					edge := lockOrder(id, nb)
					bottles[edge].Lock()
				}

				p.Record(core.ComSede, time.Since(thirstyStart))

				// ── 3. BEBENDO ──
				drinkStart := time.Now()
				time.Sleep(1 * time.Second) // bebe por 1 segundo
				p.Record(core.Bebendo, time.Since(drinkStart))

				// --- liberar garrafas ---
				for _, nb := range chosen {
					edge := lockOrder(id, nb)
					bottles[edge].Unlock()
				}

				arb.release() // libera vaga no semáforo
			}
		}(id)
	}

	wg.Wait()
	return philosophers
}

// lockOrder devolve o par ordenado {min,max} para usar como chave do mapa.
func lockOrder(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}
