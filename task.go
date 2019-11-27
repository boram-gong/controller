package controller

import (
	"fmt"
	"sync"
	"time"
)

type goTask struct {
	sync.RWMutex
	TakeName  string
	TaskOrder string
	Cyclic  bool
	Step      int
	Status    string
	DownChan  chan int
}

func new_GoTask(task_name string, task_order string, cyclic bool, step int) (go_task *goTask) {
	go_task = new(goTask)
	go_task.TakeName = task_name
	go_task.TaskOrder = task_order
	go_task.Cyclic = cyclic
	go_task.Step = step
	go_task.Status = "stop(first status)"
	go_task.DownChan = make(chan int, 1)
	return
}

func (this *goTask) start() {
	this.Status = "start"
	go func() {
		for {
			if !this.Cyclic {
				out, err := cmdWork(this.TaskOrder)
				if err != nil {
					fmt.Println(fmt.Sprintf("%s(%s): stop, error: %v", this.TakeName, this.TaskOrder, err))
					this.Status = "fail(order err)"
					return
				} else {
					this.Status = "successful"
					outDataChan <- outStringDeal(out)
					return
				}
			}
			select {
			case <-time.Tick(time.Duration(this.Step) * time.Second):
				out, err := cmdWork(this.TaskOrder)
				if err != nil {
					fmt.Println(fmt.Sprintf("%s(%s): stop, error: %v", this.TakeName, this.TaskOrder, err))
					this.Status = "stop(order err)"
					return
				}
				outDataChan <- outStringDeal(out)
			case <-this.DownChan:
				return
			}
		}
	}()
}

func (this *goTask) stop() {
	if !this.Cyclic {
		return
	}
	this.DownChan <- 1
	this.Lock()
	this.Status = "stop(initiative stop)"
	this.Unlock()
}
