package solution1_ordering

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/William-SWS/concurrent-programming-work/pkg/bottle"
	"github.com/William-SWS/concurrent-programming-work/pkg/philosopher"
)

type OrderingPhilosopher struct {
	Base *philosopher.Philosopher
}

func (p *OrderingPhilosopher) Run(rounds int) {

	for i := 0; i < rounds; i++ {

		p.think()

		selected := p.chooseBottles()

		p.acquireBottles(selected)

		p.drink()

		p.releaseBottles(selected)
	}

	fmt.Printf(
		"Philosopher %d finished execution\n",
		p.Base.ID,
	)
}

func (p *OrderingPhilosopher) think() {

	p.Base.State = philosopher.Tranquilo

	start := time.Now()

	fmt.Printf(
		"[TRANQUILO] Philosopher %d (degree %d)\n",
		p.Base.ID,
		p.Base.Degree,
	)

	// Thinking time: 0 to n (degree) seconds
	duration := time.Duration(rand.Intn(p.Base.Degree+1)) * time.Second
	time.Sleep(duration)

	p.Base.ThinkingTime += time.Since(start)
}

func (p *OrderingPhilosopher) chooseBottles() []*bottle.Bottle {
	total := len(p.Base.Bottles)
	if total < 2 {
		return p.Base.Bottles
	}

	// Choose count between 2 and n (total bottles)
	count := rand.Intn(total-1) + 2

	shuffled := make([]*bottle.Bottle, total)
	copy(shuffled, p.Base.Bottles)
	rand.Shuffle(total, func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

func (p *OrderingPhilosopher) acquireBottles(
	bottles []*bottle.Bottle,
) {

	p.Base.State = philosopher.ComSede

	start := time.Now()

	// GLOBAL RESOURCE ORDERING
	sort.Slice(bottles, func(i, j int) bool {
		return bottles[i].ID < bottles[j].ID
	})

	fmt.Printf(
		"[COM SEDE] Philosopher %d waiting %d resources\n",
		p.Base.ID,
		len(bottles),
	)

	for _, b := range bottles {
		b.Mutex.Lock()
	}

	p.Base.ThirstyTime += time.Since(start)
}

func (p *OrderingPhilosopher) drink() {

	p.Base.State = philosopher.Bebendo

	start := time.Now()

	fmt.Printf(
		"[BEBENDO] Philosopher %d\n",
		p.Base.ID,
	)

	// Drinking time: 1 second
	time.Sleep(1 * time.Second)

	p.Base.DrinkingTime += time.Since(start)

	p.Base.DrinkCount++
}

func (p *OrderingPhilosopher) releaseBottles(
	bottles []*bottle.Bottle,
) {

	for _, b := range bottles {

		b.Mutex.Unlock()

		fmt.Printf(
			"Philosopher %d released bottle %d\n",
			p.Base.ID,
			b.ID,
		)
	}
}
