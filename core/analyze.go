package core

import (
	"fmt"
	"io"
	"math"
	"sort"
	"time"
)

type starvationInfo struct {
	philosopherID int
	degree        int
	wait          time.Duration
	groupMean     time.Duration
}

// analyzeStarvation examina os tempos de espera (com_sede) e identifica
// filósofos com espera desproporcional em relação ao grupo de mesmo grau.
// Critérios:
//   - outlier: espera > 2× a média do grupo de mesmo grau
//   - severo: espera > 3× a média do grupo
//   - também calcula a razão max/mean global como métrica de fairness
func analyzeStarvation(phs []*Philosopher) ([]starvationInfo, time.Duration, time.Duration, float64) {
	esperaPorGrau := map[int][]time.Duration{}
	mediaPorGrau := map[int]time.Duration{}
	for _, p := range phs {
		esperaPorGrau[p.Degree] = append(esperaPorGrau[p.Degree], p.Metrics.ComSede)
	}
	for g, ds := range esperaPorGrau {
		mediaPorGrau[g] = media(ds)
	}

	var maxWait time.Duration
	var totalWait time.Duration
	flags := []starvationInfo{}

	for _, p := range phs {
		gm := mediaPorGrau[p.Degree]
		w := p.Metrics.ComSede
		if w > maxWait {
			maxWait = w
		}
		totalWait += w
		if gm > 0 && w > 2*gm {
			flags = append(flags, starvationInfo{
				philosopherID: p.ID,
				degree:        p.Degree,
				wait:          w,
				groupMean:     gm,
			})
		}
	}

	meanWait := totalWait / time.Duration(len(phs))
	ratio := 0.0
	if meanWait > 0 {
		ratio = maxWait.Seconds() / meanWait.Seconds()
	}
	return flags, maxWait, meanWait, ratio
}

// starvationReport escreve a seção de análise de starvation no relatório.
func starvationReport(w io.Writer, phs []*Philosopher) {
	flags, maxWait, meanWait, ratio := analyzeStarvation(phs)

	fmt.Fprintln(w, "# starvation analysis:")

	var allWaits []float64
	for _, p := range phs {
		allWaits = append(allWaits, p.Metrics.ComSede.Seconds())
	}
	sort.Float64s(allWaits)
	stddev := stddevFloat(allWaits)
	cv := 0.0
	if meanWait > 0 {
		cv = stddev / meanWait.Seconds()
	}
	fmt.Fprintf(w, "#   max_wait=%.2fs mean_wait=%.2fs stddev=%.2fs max/mean=%.2f cv=%.2f\n",
		maxWait.Seconds(), meanWait.Seconds(), stddev, ratio, cv)

	esperaPorGrau := map[int][]time.Duration{}
	for _, p := range phs {
		esperaPorGrau[p.Degree] = append(esperaPorGrau[p.Degree], p.Metrics.ComSede)
	}
	graus := make([]int, 0, len(esperaPorGrau))
	for g := range esperaPorGrau {
		graus = append(graus, g)
	}
	sort.Ints(graus)

	for _, g := range graus {
		ds := esperaPorGrau[g]
		gm := media(ds)
		var gMax time.Duration
		for _, d := range ds {
			if d > gMax {
				gMax = d
			}
		}
		gRatio := 0.0
		if gm > 0 {
			gRatio = gMax.Seconds() / gm.Seconds()
		}
		fmt.Fprintf(w, "#   grau=%d media=%.2fs max=%.2fs max/media=%.2f (n=%d)\n",
			g, gm.Seconds(), gMax.Seconds(), gRatio, len(ds))
	}

	if len(flags) == 0 {
		fmt.Fprintln(w, "#   resultado: nenhum filosofo com espera acima de 2x a media do grupo")
	} else {
		fmt.Fprintln(w, "#   ALERTA: filosofos com espera >2x a media do grupo de mesmo grau:")
		for _, f := range flags {
			pct := (f.wait.Seconds() / f.groupMean.Seconds()) * 100
			severity := "outlier"
			if f.wait > 3*f.groupMean {
				severity = "CRITICO"
			}
			fmt.Fprintf(w, "#     filo %d (grau %d): esperou %.2fs (%.0f%% da media %.2fs) [%s]\n",
				f.philosopherID, f.degree, f.wait.Seconds(), pct, f.groupMean.Seconds(), severity)
		}
	}
}

func stddevFloat(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	mean := sum / float64(len(vals))
	var sqSum float64
	for _, v := range vals {
		d := v - mean
		sqSum += d * d
	}
	return math.Sqrt(sqSum / float64(len(vals)))
}
