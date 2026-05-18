package philosopher

import (
	"time"

	"github.com/William-SWS/concurrent-programming-work/pkg/bottle"
)

type Philosopher struct {
	ID int

	Bottles []*bottle.Bottle

	State State

	Degree int

	DrinkCount int

	ThinkingTime time.Duration
	ThirstyTime  time.Duration
	DrinkingTime time.Duration
}