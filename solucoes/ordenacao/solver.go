// Package ordenacao implementa a Solucao 1: numeracao das garrafas
// (ordenacao de recursos). Cada garrafa recebe uma ordem global, e todos os
// filosofos adquirem recursos nessa mesma ordem para evitar espera circular.
package ordenacao

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/William-SWS/concurrent-programming-work/core"
)

type Solver struct{}

// New devolve o solver desta solucao.
func New() core.Solver { return &Solver{} }

func (s *Solver) Name() string { return "Ordenacao de Recursos" }

func (s *Solver) Run(g *core.Graph, rounds int) []*core.Philosopher {
	philosophers := make([]*core.Philosopher, g.N)
	for i := 0; i < g.N; i++ {
		philosophers[i] = core.NewPhilosopher(i, g)
	}

	bottles := make(map[[2]int]*sync.Mutex)
	order := make(map[[2]int]int)
	for id, edge := range g.Edges() {
		bottles[edge] = &sync.Mutex{}
		order[edge] = id
	}

	var wg sync.WaitGroup
	for id := 0; id < g.N; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runPhilosopher(philosophers[id], g, bottles, order, rounds)
		}(id)
	}

	wg.Wait()
	return philosophers
}

func runPhilosopher(
	p *core.Philosopher,
	g *core.Graph,
	bottles map[[2]int]*sync.Mutex,
	order map[[2]int]int,
	rounds int,
) {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(p.ID)))
	neighbors := g.Neighbors(p.ID)
	degree := len(neighbors)

	for i := 0; i < rounds; i++ {
		thinkingTime := time.Duration(r.Intn(degree+1)) * time.Second
		thinkStart := time.Now()
		time.Sleep(thinkingTime)
		p.Record(core.Tranquilo, time.Since(thinkStart))

		thirstyStart := time.Now()
		chosen := chooseNeighbors(r, neighbors)
		sort.Slice(chosen, func(i, j int) bool {
			return order[lockOrder(p.ID, chosen[i])] < order[lockOrder(p.ID, chosen[j])]
		})

		for _, nb := range chosen {
			bottles[lockOrder(p.ID, nb)].Lock()
		}
		p.Record(core.ComSede, time.Since(thirstyStart))

		drinkStart := time.Now()
		time.Sleep(1 * time.Second)
		p.Record(core.Bebendo, time.Since(drinkStart))

		for _, nb := range chosen {
			bottles[lockOrder(p.ID, nb)].Unlock()
		}
	}
}

func chooseNeighbors(r *rand.Rand, neighbors []int) []int {
	total := len(neighbors)
	if total < 2 {
		chosen := make([]int, total)
		copy(chosen, neighbors)
		return chosen
	}

	count := r.Intn(total-1) + 2
	shuffled := make([]int, total)
	copy(shuffled, neighbors)
	r.Shuffle(total, func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:count]
}

func lockOrder(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}
