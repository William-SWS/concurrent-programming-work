package core

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// Metrics guarda o tempo acumulado de um filósofo em cada estado
type Metrics struct {
	Tranquilo time.Duration
	ComSede   time.Duration // tempo de espera (base para avaliar starvation)
	Bebendo   time.Duration
	Drinks    int
}

// Report imprime a tabela por filósofo + o tempo total e a espera média
// agrupada por grau.
// O formato é estável (TSV) de propósito, para o scripts/compare_results.py
// conseguir parsear sem ambiguidade.
func Report(w io.Writer, solucao string, total time.Duration, phs []*Philosopher) {
	fmt.Fprintf(w, "# solucao=%s\n", solucao)
	fmt.Fprintf(w, "# tempo_total=%.2fs\n", total.Seconds())
	fmt.Fprintln(w, "id\tgrau\ttranquilo\tcom_sede\tbebendo\tbebidas")

	esperaPorGrau := map[int][]time.Duration{}
	for _, p := range phs {
		m := p.Metrics
		fmt.Fprintf(w, "%d\t%d\t%.2f\t%.2f\t%.2f\t%d\n",
			p.ID, p.Degree, m.Tranquilo.Seconds(), m.ComSede.Seconds(), m.Bebendo.Seconds(), m.Drinks)
		esperaPorGrau[p.Degree] = append(esperaPorGrau[p.Degree], m.ComSede)
	}

	fmt.Fprintln(w, "# espera media (com sede) por grau:")
	graus := make([]int, 0, len(esperaPorGrau))
	for g := range esperaPorGrau {
		graus = append(graus, g)
	}
	sort.Ints(graus)
	for _, g := range graus {
		fmt.Fprintf(w, "# grau=%d media=%.2fs (n=%d)\n", g, media(esperaPorGrau[g]).Seconds(), len(esperaPorGrau[g]))
	}
}

func media(ds []time.Duration) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	var soma time.Duration
	for _, d := range ds {
		soma += d
	}
	return soma / time.Duration(len(ds))
}
