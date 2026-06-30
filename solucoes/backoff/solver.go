package backoff

import (
	"math/rand"
	"sync"
	"time"

	"github.com/William-SWS/concurrent-programming-work/core"
)

// BackoffSolver implementa a Solução 4 - Aleatoriedade (Randomized Backoff)
type BackoffSolver struct{}

// New instancia a solução para os testes e para o runner principal
func New() core.Solver {
	return &BackoffSolver{}
}

// Name retorna o rótulo da solução, conforme exigido pela interface core.Solver
func (b *BackoffSolver) Name() string {
	return "backoff"
}

// Run executa a simulação com base no Grafo.
func (b *BackoffSolver) Run(g *core.Graph, totalBebidas int) []*core.Philosopher {
	// 1. Mapear as garrafas (Mutexes)
	// Como g.Edges() devolve [][2]int, usamos o array [2]int como chave do mapa
	garrafas := make(map[[2]int]*sync.Mutex)
	for _, edge := range g.Edges() {
		garrafas[edge] = &sync.Mutex{}
	}

	// 2. Instanciar os Filósofos usando o construtor fornecido pelo pacote core
	var filosofos []*core.Philosopher
	for i := 0; i < g.N; i++ {
		filosofos = append(filosofos, core.NewPhilosopher(i, g))
	}

	var wg sync.WaitGroup

	// 3. Iniciar as Goroutines (as threads dos processos concorrentes)
	for _, f := range filosofos {
		wg.Add(1)
		go rotinaFilosofo(f, g, garrafas, totalBebidas, &wg)
	}

	// 4. Aguardar o fim da simulação de todos os filósofos
	wg.Wait()
	return filosofos
}

// rotinaFilosofo contem a logica de Randomized Backoff 
func rotinaFilosofo(f *core.Philosopher, g *core.Graph, garrafas map[[2]int]*sync.Mutex, totalBebidas int, wg *sync.WaitGroup) {
	defer wg.Done()
	
	// Cria uma fonte de aleatoriedade exclusiva para essa goroutine
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(f.ID)))
	
	// Puxa a lista de vizinhos e a quantidade de arestas 
	vizinhos := g.Neighbors(f.ID) 
	numVizinhos := len(vizinhos)

	for i := 0; i < totalBebidas; i++ {
		// --- 1. ESTADO: TRANQUILO ---
		// De 0 até n segundos, conforme as regras do trabalho
		tTranquilo := time.Duration(r.Float64() * float64(numVizinhos) * float64(time.Second))
		time.Sleep(tTranquilo)
		f.Record(core.Tranquilo, tTranquilo) 

		// --- 2. ESTADO: COM SEDE ---
		inicioSede := time.Now()

		// Sorteia quantas garrafas pegar (de 2 até n)
		qtdGarrafas := numVizinhos
		if numVizinhos >= 2 {
			qtdGarrafas = r.Intn(numVizinhos-1) + 2
		}

		// Sorteia aleatoriamente QUAIS vizinhos ele vai pedir a garrafa
		vizinhosEmbaralhados := make([]int, len(vizinhos))
		copy(vizinhosEmbaralhados, vizinhos)
		r.Shuffle(len(vizinhosEmbaralhados), func(i, j int) {
			vizinhosEmbaralhados[i], vizinhosEmbaralhados[j] = vizinhosEmbaralhados[j], vizinhosEmbaralhados[i]
		})
		vizinhosEscolhidos := vizinhosEmbaralhados[:qtdGarrafas]

		// LÓGICA DO RANDOMIZED BACKOFF 
		for {
			sucesso := true
			var garrafasPegas []*sync.Mutex

			for _, v := range vizinhosEscolhidos {
				menor, maior := f.ID, v
				if f.ID > v {
					menor, maior = v, f.ID
				}
				
				// Busca a garrafa pela chave padronizada
				m := garrafas[[2]int{menor, maior}]
				
				// TryLock tenta travar a garrafa sem bloquear a thread
				if m.TryLock() {
					garrafasPegas = append(garrafasPegas, m)
				} else {
					// Se alguém ja pegou a garrafa, marcamos falha e paramos de tentar
					sucesso = false
					break
				}
			}

			if sucesso {
				break // Conseguiu pegar tudo, vai sair do loop e beber
			} else {
				// Falhou: Solta imediatamente as garrafas que ja tinha travado
				for _, m := range garrafasPegas {
					m.Unlock()
				}
				// Backoff: dorme um tempinho aleatório (ex: até 0.1s) para desincronizar dos vizinhos
				time.Sleep(time.Duration(r.Float64() * 0.1 * float64(time.Second)))
			}
		}

		// Registra o tempo total que ficou preso no ciclo "com sede"
		f.Record(core.ComSede, time.Since(inicioSede))

		// --- 3. ESTADO: BEBENDO ---
		// Bebendo leva exato 1 segundo, conforme regras
		time.Sleep(1 * time.Second)
		
		// O Record de Bebendo 
		f.Record(core.Bebendo, 1*time.Second)

		// Após beber, libera todas as garrafas para que os vizinhos possam usar
		for _, v := range vizinhosEscolhidos {
			menor, maior := f.ID, v
			if f.ID > v {
				menor, maior = v, f.ID
			}
			garrafas[[2]int{menor, maior}].Unlock()
		}
	}
}