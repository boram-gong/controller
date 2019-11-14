package controller

import (
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

var outDataChan = make(chan string, 100)
var closeMessagePoolChan = make(chan int, 1)

func ResetMessagePool(length int) {
	outDataChan = nil
	outDataChan = make(chan string, length)
}

func CloseMessagePoolChan() {
	closeMessagePoolChan <- 1
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
	this.Lock()
	defer this.Unlock()
	task, ok := this.Pool[task_name]
	if ok {
		task.stop()
	}
	delete(this.Pool, task_name)
}

func (this *GoPool) DeleteAllTake() {
	this.Lock()
	defer this.Unlock()
	for _, task := range this.Pool {
		task.stop()
	}
	this.Pool = nil
	this.Pool = make(map[string]*goTask)
}

func (this *GoPool) StartAllTake() {
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

func (this *GoPool) StopAllTake() {
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

func (this *GoPool) WaitDataFromMessagePool() (data string, err error) {
	select {
	case data = <-outDataChan:
		return
	case <-closeMessagePoolChan:
		return "", errors.New("close")
	}
}

func (this *GoPool) QueryTaskState(task_name string) string {
	this.RLock()
	defer this.RUnlock()
	task, ok := this.Pool[task_name]
	if ok {
		return task.Status
	} else {
		return "don't have this task"
	}
}

func (this *GoPool) QueryAllTaskState() (result string) {
	this.RLock()
	defer this.RUnlock()
	if len(this.Pool) == 0 {
		return "no task"
	}
	for name, task := range this.Pool {
		result += fmt.Sprintf("%s: %s\n", name, task.Status)
	}
	return
}
