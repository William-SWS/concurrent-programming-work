// Command runner executa uma solução sobre um grafo e imprime o relatório.
//
// Uso:
//
//	go run ./cmd/runner -solucao=ordenacao -grafo=data/caso2_bar_6.txt -rodadas=6
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/William-SWS/concurrent-programming-work/core"
	"github.com/William-SWS/concurrent-programming-work/solucoes/arbitro"
	"github.com/William-SWS/concurrent-programming-work/solucoes/backoff"
	"github.com/William-SWS/concurrent-programming-work/solucoes/chandy_misra"
	"github.com/William-SWS/concurrent-programming-work/solucoes/ordenacao"
)

// registro mapeia o nome da solução (flag) para seu construtor.
var registro = map[string]func() core.Solver{
	"ordenacao":    ordenacao.New,
	"arbitro":      arbitro.New,
	"chandy_misra": chandy_misra.New,
	"backoff":      backoff.New,
}

func main() {
	nome := flag.String("solucao", "", "solução a rodar: "+nomesDisponiveis())
	grafo := flag.String("grafo", "", "caminho do .txt com a matriz de adjacência")
	rodadas := flag.Int("rodadas", 6, "quantas vezes cada filósofo deve beber")
	flag.Parse()

	novo, ok := registro[*nome]
	if !ok {
		fmt.Fprintf(os.Stderr, "solução inválida %q. disponíveis: %s\n", *nome, nomesDisponiveis())
		os.Exit(2)
	}
	if *grafo == "" {
		fmt.Fprintln(os.Stderr, "informe -grafo")
		os.Exit(2)
	}

	g, err := core.LoadGraph(*grafo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	solver := novo()
	inicio := time.Now()
	phs := solver.Run(g, *rodadas)
	total := time.Since(inicio)

	core.Report(os.Stdout, solver.Name(), total, phs)
}

func nomesDisponiveis() string {
	ns := make([]string, 0, len(registro))
	for n := range registro {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	return fmt.Sprint(ns)
}
