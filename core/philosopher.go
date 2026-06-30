package core

import "time"

// Philosopher é a abstração compartilhada por todas as soluções. Cada solução
// move o filósofo pelos estados à sua maneira, mas todas reportam tempo via os
// mesmos campos para que a comparação final seja justa.
type Philosopher struct {
	ID     int
	Degree int // número de arestas adjacentes (n)

	Metrics Metrics
}

// NewPhilosopher cria um filósofo associado ao vértice id do grafo g.
func NewPhilosopher(id int, g *Graph) *Philosopher {
	return &Philosopher{ID: id, Degree: g.Degree(id)}
}

// Record acumula o tempo gasto em um estado. Cada solução deve cronometrar
// suas transições e chamar este método (ex.: medir quanto ficou ComSede até
// conseguir todas as garrafas).
func (p *Philosopher) Record(s State, d time.Duration) {
	switch s {
	case Tranquilo:
		p.Metrics.Tranquilo += d
	case ComSede:
		p.Metrics.ComSede += d
	case Bebendo:
		p.Metrics.Bebendo += d
		p.Metrics.Drinks++
	}
}
