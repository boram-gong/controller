package controller

import (
	"fmt"
	"testing"
	"time"
)

func demo(i int) int {
	return i
}

func TestPool(t *testing.T) {
	pool := NewPool()
	pool.AppendCyclicFuncTake("taskName", demo, ArgsMaker(1), true, 1, nil)
	go func() {
		for r := range pool.OutDataChan {
			fmt.Println(r)
		}
	}()
	pool.StartAllTake()
	time.Sleep(3 * time.Second)
	fmt.Println(pool.QueryTaskMessage("taskName"))
	time.Sleep(2 * time.Second)
	pool.StopAllTake()
	fmt.Println(pool.QueryTaskMessage("taskName"))
}
