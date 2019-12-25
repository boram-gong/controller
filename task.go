package controller

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

const (
	INIT    = "Init"
	START   = "Start"
	SUCCESS = "Successful"
	NO_TASK = "NoTask"
	FAIL    = "Fail"
	SERIOUS = "SeriousErr"
	STOP    = "Stop"
)

type goTask struct {
	sync.RWMutex
	TakeName  string
	TaskOrder string
	TaskArgs  []interface{}
	TaskFunc  interface{}
	Cyclic    bool
	Step      int
	Status    string
	Note      interface{}
	DownChan  chan int
}

func new_OrderTask(task_name string, task_order string, cyclic bool, step int, note interface{}) (go_task *goTask) {
	go_task = new(goTask)
	go_task.TakeName = task_name
	go_task.TaskOrder = task_order
	go_task.Cyclic = cyclic
	go_task.Step = step
	go_task.Status = INIT
	go_task.DownChan = make(chan int, 1)
	go_task.Note = note
	return
}

func new_FuncTask(task_name string, task_func interface{}, task_args []interface{}, cyclic bool, step int, note interface{}) (go_task *goTask) {
	go_task = new(goTask)
	go_task.TakeName = task_name
	go_task.TaskFunc = task_func
	go_task.TaskArgs = task_args
	go_task.Cyclic = cyclic
	go_task.Step = step
	go_task.Status = INIT
	go_task.DownChan = make(chan int, 1)
	go_task.Note = note
	return
}

func (this *goTask) start(ch chan []interface{}, stat chan []string) {
	if this.Status == START {
		LogChan <- fmt.Sprintf("Warn: '%v' task is running, don't to need run again", this.TakeName)
		return
	}
	this.Status = START
	stat <- []string{this.TakeName, this.Status}
	go func() {
		for {
			if !this.Cyclic {
				if this.run(ch, stat) {
					if this.Status != SERIOUS && this.Status != NO_TASK {
						this.Status = SUCCESS + "_" + time.Now().Format("2006/01/02/15:04")
						stat <- []string{this.TakeName, this.Status}
					}
				}
				return
			}
			select {
			case <-time.Tick(time.Duration(this.Step) * time.Second):
				if !this.run(ch, stat) {
					LogChan <- fmt.Sprintf("Error: '%v' task running exception", this.TakeName)
					return
				}
			case <-this.DownChan:
				LogChan <- fmt.Sprintf("Info: '%v' task manual stop", this.TakeName)
				return
			}
		}
	}()
}

func (this *goTask) stop(stat chan []string) {
	if !this.Cyclic {
		return
	}
	if this.Status == STOP {
		return
	}
	this.DownChan <- 1
	this.Lock()
	this.Status = STOP
	this.Unlock()
	stat <- []string{this.TakeName, this.Status}
}

func (this *goTask) run(ch chan []interface{}, stat chan []string) bool {
	var (
		out    string
		err    error
		result []interface{}
	)
	if this.TaskOrder != "" {
		if out, err = cmdWork(this.TaskOrder); err != nil {
			LogChan <- fmt.Sprintf("Error: %s(%s) stop, error: %v", this.TakeName, this.TaskOrder, err)
			this.Status = FAIL + "_" + time.Now().Format("2006/01/02/15:04")
			stat <- []string{this.TakeName, this.Status}
			return false
		} else {
			if outStringDeal(out) == "" {
				return true
			}
			r := []interface{}{}
			r = append(r, this.TakeName, outStringDeal(out))
			ch <- r
			return true
		}
	} else {
		if this.TaskFunc == nil {
			LogChan <- fmt.Sprintf("Error: %s no task", this.TakeName)
			this.Status = NO_TASK
			stat <- []string{this.TakeName, this.Status}
			return false
		} else {
			result = this.funGenerator()
			if this.Status == SERIOUS || this.Status == NO_TASK {
				stat <- []string{this.TakeName, this.Status}
				return false
			}
			if len(result) != 0 {
				ch <- result
			}
			return true
		}
	}
}

func (this *goTask) funGenerator() []interface{} {
	defer func(g *goTask) {
		serious_err := recover()
		if serious_err != nil {
			LogChan <- fmt.Sprintf("Error: %s task fail: %v", this.TakeName, serious_err)
			g.Status = SERIOUS
			if g.Cyclic {
				g.DownChan <- 1
			}
		}
	}(this)
	var (
		rFun              reflect.Value
		args, returnValue []reflect.Value
		l                 []interface{}
	)
	rFun = reflect.ValueOf(this.TaskFunc)
	if rFun.Kind() != reflect.Func {
		LogChan <- fmt.Sprintf("Error: %s no task", this.TakeName)
		this.Status = NO_TASK
		return l
	}
	for _, a := range this.TaskArgs {
		args = append(args, reflect.ValueOf(a))
	}
	returnValue = rFun.Call(args)
	l = append(l, this.TakeName)
	for _, r := range returnValue {
		l = append(l, r.Interface())
	}
	return l
}
