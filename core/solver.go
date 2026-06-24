package core

// Solver é o contrato que cada uma das 4 soluções implementa. É o que permite
// dividir o trabalho entre 4 pessoas sem conflito: todos recebem o mesmo grafo
// e o mesmo número de rodadas, e devolvem os filósofos com as métricas preenchidas.
type Solver interface {
	// Name é o rótulo da solução
	Name() string

	// Run executa a simulação até cada filósofo beber 'rounds' vezes e devolve
	// os filósofos com Metrics preenchidas. Não pode ocorrer deadlock.
	Run(g *Graph, rounds int) []*Philosopher
}
