package bottle

import "sync"

type Bottle struct {
	ID    int
	Mutex sync.Mutex
}