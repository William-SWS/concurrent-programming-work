package solution1_ordering

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/William-SWS/concurrent-programming-work/pkg/graph"
)

type Solver struct {
	Graph     *graph.Graph
	Rounds    int
	CaseID    int
	StartTime time.Time
	EndTime   time.Time
}

func (s *Solver) Run() {
	s.StartTime = time.Now()
	var wg sync.WaitGroup

	for _, p := range s.Graph.Philosophers {
		op := &OrderingPhilosopher{
			Base: p,
		}
		wg.Add(1)
		go func(ph *OrderingPhilosopher) {
			defer wg.Done()
			ph.Run(s.Rounds)
		}(op)
	}

	wg.Wait()
	s.EndTime = time.Now()

	s.printResults()
	s.saveResults()
}

func (s *Solver) printResults() {
	fmt.Println("\n=== RESULTS ===")
	totalDuration := s.EndTime.Sub(s.StartTime)
	fmt.Printf("Total execution time: %v\n\n", totalDuration)

	for _, p := range s.Graph.Philosophers {
		fmt.Printf("Philosopher %d (Degree %d)\n", p.ID, p.Degree)
		fmt.Printf("  Tranquilo: %v\n", p.ThinkingTime)
		fmt.Printf("  Com Sede:  %v\n", p.ThirstyTime)
		fmt.Printf("  Bebendo:   %v\n", p.DrinkingTime)
		fmt.Printf("  Total Drinks: %d\n\n", p.DrinkCount)
	}
}

type PhilosopherResult struct {
	ID           int           `json:"id"`
	Degree       int           `json:"degree"`
	ThinkingTime time.Duration `json:"thinking_time_ns"`
	ThirstyTime  time.Duration `json:"thirsty_time_ns"`
	DrinkingTime time.Duration `json:"drinking_time_ns"`
	DrinkCount   int           `json:"drink_count"`
}

type DegreeMetric struct {
	Degree          int           `json:"degree"`
	Count           int           `json:"count"`
	AvgThirstyTime  time.Duration `json:"avg_thirsty_time_ns"`
}

type SimulationResult struct {
	CaseID        int                 `json:"case_id"`
	TotalDuration time.Duration       `json:"total_duration_ns"`
	Philosophers  []PhilosopherResult `json:"philosophers"`
	DegreeMetrics []DegreeMetric      `json:"degree_metrics"`
}

func (s *Solver) saveResults() {
	resDir := fmt.Sprintf("results/caso%d", s.CaseID)
	os.MkdirAll(resDir, 0755)

	simRes := SimulationResult{
		CaseID:        s.CaseID,
		TotalDuration: s.EndTime.Sub(s.StartTime),
	}

	degreeThirsty := make(map[int][]time.Duration)

	for _, p := range s.Graph.Philosophers {
		pr := PhilosopherResult{
			ID:           p.ID,
			Degree:       p.Degree,
			ThinkingTime: p.ThinkingTime,
			ThirstyTime:  p.ThirstyTime,
			DrinkingTime: p.DrinkingTime,
			DrinkCount:   p.DrinkCount,
		}
		simRes.Philosophers = append(simRes.Philosophers, pr)
		degreeThirsty[p.Degree] = append(degreeThirsty[p.Degree], p.ThirstyTime)
	}

	for deg, times := range degreeThirsty {
		var total time.Duration
		for _, t := range times {
			total += t
		}
		simRes.DegreeMetrics = append(simRes.DegreeMetrics, DegreeMetric{
			Degree:         deg,
			Count:          len(times),
			AvgThirstyTime: total / time.Duration(len(times)),
		})
	}

	// Save JSON
	jsonFile := filepath.Join(resDir, "results.json")
	jsonData, _ := json.MarshalIndent(simRes, "", "  ")
	os.WriteFile(jsonFile, jsonData, 0644)

	// Save TXT
	txtFile := filepath.Join(resDir, "results.txt")
	f, _ := os.Create(txtFile)
	defer f.Close()

	fmt.Fprintf(f, "=== Simulation Results - Case %d ===\n", s.CaseID)
	fmt.Fprintf(f, "Total duration: %v\n\n", simRes.TotalDuration)
	fmt.Fprintf(f, "--- Metrics by Degree ---\n")
	for _, m := range simRes.DegreeMetrics {
		fmt.Fprintf(f, "Degree %d: Count=%d, Avg Waiting (Thirsty) Time=%v\n", m.Degree, m.Count, m.AvgThirstyTime)
	}
	fmt.Fprintf(f, "\n--- Individual Results ---\n")
	for _, p := range simRes.Philosophers {
		fmt.Fprintf(f, "Philosopher %d (Deg %d):\n", p.ID, p.Degree)
		fmt.Fprintf(f, "  Tranquilo: %v\n", p.ThinkingTime)
		fmt.Fprintf(f, "  Com Sede:  %v\n", p.ThirstyTime)
		fmt.Fprintf(f, "  Bebendo:   %v\n", p.DrinkingTime)
	}
}
