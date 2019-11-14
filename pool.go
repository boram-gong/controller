package controller

import (
	"errors"
	"fmt"
	"sync"
)

var outDataChan = make(chan string, 100)

func ResetMessagePool(length int) {
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
	this.StopAllTake()
	this.Lock()
	defer this.Unlock()
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
	if len(this.Pool) == 0 {
		return "", errors.New("No task currently exists")
	}
	if !this.queryAllStop() {
		return "", errors.New("all task stop")
	}
	select {
	case data = <-outDataChan:
		return
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

func (this *GoPool) queryAllStop() bool {
	this.RLock()
	defer this.RUnlock()
	for _, task := range this.Pool {
		if task.Status == "start" {
			return true
		}
	}
	return false
}

func (this *GoPool) QueryAllTaskState() (result string) {
	this.RLock()
	defer this.RUnlock()
	for name, task := range this.Pool {
		result += fmt.Sprintf("%s: %s\n", name, task.Status)
	}
	return
}
