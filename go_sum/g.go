package go_sum

import (
	"sync"
)

type GoroutineSum struct {
	Size chan int
	wg   *sync.WaitGroup
}

func New(size int) *GoroutineSum {
	if size <= 0 {
		size = 1
	}
	return &GoroutineSum{
		Size: make(chan int, size),
		wg:   &sync.WaitGroup{},
	}
}

func (gs *GoroutineSum) Add(delta int) {
	if delta < 1 {
		delta = 1
	}
	for i := 0; i < delta; i++ {
		gs.Size <- 1
	}
	gs.wg.Add(delta)
}

func (gs *GoroutineSum) Done() {
	<-gs.Size
	gs.wg.Done()
}

func (gs *GoroutineSum) Wait() {
	gs.wg.Wait()
}
