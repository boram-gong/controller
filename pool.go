package controller

import (
	"sync"
	"time"
)

var outDataChan = make(chan string, 100)

func ResetOutDataChan(length int) {
	outDataChan = nil
	outDataChan = make(chan string, length)
}

type GoPool struct {
	sync.RWMutex
	Pool map[string]*goTask
}

func NewPool() (pool *GoPool) {
	pool = new(GoPool)
	pool.Pool = make(map[string]*goTask)
	return
}

func (this *GoPool) AppendTake(task_name string, task_order string, step int) {
	task := new_GoTask(task_name, task_order, step)
	this.Lock()
	defer this.Unlock()
	this.Pool[task_name] = task
}

func (this *GoPool) DeleteTake(task_name string) {
	this.StopOneTask(task_name)
	this.Lock()
	defer this.Unlock()
	delete(this.Pool, task_name)
}

func (this *GoPool) DeleteAllTake() {
	this.StopAll()
	this.Lock()
	defer this.Unlock()
	this.Pool = nil
	this.Pool = make(map[string]*goTask)
}

func (this *GoPool) StartAll() {
	this.RLock()
	defer this.RUnlock()
	for _, task := range this.Pool {
		task.start()
	}
}

func (this *GoPool) StartOneTask(task_name string) bool {
	this.RLock()
	defer this.RUnlock()
	task, ok := this.Pool[task_name]
	if !ok {
		return false
	} else {
		task.start()
		return true
	}
}

func (this *GoPool) StopAll() {
	this.RLock()
	defer this.RUnlock()
	for _, task := range this.Pool {
		task.stop()
	}
}

func (this *GoPool) StopOneTask(task_name string) bool {
	this.RLock()
	defer this.RUnlock()
	task, ok := this.Pool[task_name]
	if !ok {
		return false
	} else {
		task.stop()
		return true
	}
}

func (this *GoPool) WaitData() (data string) {
	select {
	case data = <-outDataChan:
		return
	}
}

type goTask struct {
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
				outDataChan <- out

			case <-this.DownChan:
				this.Status = "stop(initiative stop)"
				return

			}
		}
	}()
}

func (this *goTask) stop() {
	this.DownChan <- 1
}
