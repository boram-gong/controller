package controller

import (
	"sync"
	"time"
)

type goTask struct {
	sync.RWMutex
	TakeName  string
	TaskOrder string
	Step      int
	Status    string
	DownChan  chan int
}

func new_GoTask(task_name string, task_order string, step int) (go_task *goTask) {
	go_task = new(goTask)
	go_task.TakeName = task_name
	go_task.TaskOrder = task_order
	go_task.Step = step
	go_task.Status = "stop(first status)"
	go_task.DownChan = make(chan int, 1)
	return
}

func (this *goTask) start() {
	this.Status = "start"
	go func() {
		for {
			select {
			case <-time.Tick(time.Duration(this.Step) * time.Second):
				out, err := CmdWork(this.TaskOrder)
				if err != nil {
					this.Status = "stop(order err)"
					return
				}
				outDataChan <- OutStringDeal(out)
			case <-this.DownChan:
				return

			}
		}
	}()
}

func (this *goTask) stop() {
	this.DownChan <- 1
	this.Lock()
	this.Status = "stop(initiative stop)"
	this.Unlock()
}
