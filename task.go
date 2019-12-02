package controller

import (
	"fmt"
	"sync"
	"time"
)

const (
	INIT    = "Init"
	START   = "Start"
	SUCCESS = "Successful"
	NO_TASK = "NoTask"
	FAIL    = "Fail"
	SERIOUS = "Serious"
	STOP    = "Stop"
)

type goTask struct {
	sync.RWMutex
	TakeName  string
	TaskOrder string
	TaskArgs  []interface{}
	TaskFunc  func(args ...interface{}) (interface{}, error)
	Cyclic    bool
	Step      int
	Status    string
	DownChan  chan int
}

func new_OrderTask(task_name string, task_order string, cyclic bool, step int) (go_task *goTask) {
	go_task = new(goTask)
	go_task.TakeName = task_name
	go_task.TaskOrder = task_order
	go_task.Cyclic = cyclic
	go_task.Step = step
	go_task.Status = INIT
	go_task.DownChan = make(chan int, 1)
	return
}

func new_FuncTask(task_name string, task_func func(args ...interface{}) (interface{}, error), task_args []interface{}, cyclic bool, step int) (go_task *goTask) {
	go_task = new(goTask)
	go_task.TakeName = task_name
	go_task.TaskFunc = task_func
	go_task.TaskArgs = task_args
	go_task.Cyclic = cyclic
	go_task.Step = step
	go_task.Status = INIT
	go_task.DownChan = make(chan int, 1)
	return
}

func (this *goTask) start() {
	this.Status = START + "_" + time.Now().Format("2006/01/02/15:04")
	go func() {
		for {
			if !this.Cyclic {
				if this.run() {
					this.Status = SUCCESS + "_" + time.Now().Format("2006/01/02/15:04")
				}
				return
			}
			select {
			case <-time.Tick(time.Duration(this.Step) * time.Second):
				if !this.run() {
					return
				}
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
	this.Status = STOP + "_" + time.Now().Format("2006/01/02/15:04")
	this.Unlock()
}

func (this *goTask) run() bool {
	var (
		out    string
		err    error
		result interface{}
	)
	if this.TaskOrder != "" {
		if out, err = cmdWork(this.TaskOrder); err != nil {
			fmt.Println(fmt.Sprintf("%s(%s): stop, error: %v", this.TakeName, this.TaskOrder, err))
			this.Status = FAIL + "_" + time.Now().Format("2006/01/02/15:04")
			return false
		} else {
			outDataChan <- outStringDeal(out)
			return true
		}
	} else {
		if this.TaskFunc == nil {
			this.Status = NO_TASK
			return false
		} else {
			if result, err = this.funGenerator(); err != nil {
				fmt.Println(fmt.Sprintf("%s: fail, error: %v", this.TakeName, err))
				this.Status = FAIL + "_" + time.Now().Format("2006/01/02/15:04")
				return false
			}
			outDataChan <- result
			return true
		}
	}
}

func (this *goTask) checkFunc() {
	if recover() != nil {
		fmt.Println(this.TakeName, "serious mistake!")
		this.Status = SERIOUS + "_" + time.Now().Format("2006/01/02/15:04")
		if this.Cyclic {
			this.DownChan <- 1
		}
	}
}

func (this *goTask) funGenerator() (interface{}, error) {
	defer this.checkFunc()
	return this.TaskFunc(this.TaskArgs...)
}
