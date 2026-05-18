package graph

import (
	"github.com/William-SWS/concurrent-programming-work/pkg/bottle"
	"github.com/William-SWS/concurrent-programming-work/pkg/philosopher"
)

type Graph struct {
	Philosophers []*philosopher.Philosopher
	Bottles      []*bottle.Bottle
}

func New(matrix [][]int) *Graph {

	size := len(matrix)

	g := &Graph{}

	for i := 0; i < size; i++ {

		g.Philosophers = append(
			g.Philosophers,
			&philosopher.Philosopher{
				ID: i,
			},
		)
	}

	bottleID := 0

	for i := 0; i < size; i++ {
		degree := 0
		for j := 0; j < size; j++ {
			if matrix[i][j] == 1 {
				degree++
				if i < j {
					b := &bottle.Bottle{
						ID: bottleID,
					}

					g.Bottles =
						append(g.Bottles, b)

					g.Philosophers[i].Bottles =
						append(
							g.Philosophers[i].Bottles,
							b,
						)

					g.Philosophers[j].Bottles =
						append(
							g.Philosophers[j].Bottles,
							b,
						)

					bottleID++
				}
			}
		}
		g.Philosophers[i].Degree = degree
	}

	return g
}