package controller

import (
	"github.com/pkg/errors"
	"sync"
)

var outDataChan = make(chan interface{}, 100)
var closeMessagePoolChan = make(chan int, 1)

func ResetMessagePool(length int) {
	outDataChan = nil
	outDataChan = make(chan interface{}, length)
}

func CloseMessagePoolChan() {
	closeMessagePoolChan <- 1
}

func WaitDataFromMessagePool() (data interface{}, err error) {
	select {
	case data = <-outDataChan:
		return
	case <-closeMessagePoolChan:
		return nil, errors.New("close")
	}
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

func (this *GoPool) AppendOrderTake(task_name string, task_order string, cyclic bool, step int) {
	task := new_OrderTask(task_name, task_order, cyclic, step)
	this.Lock()
	defer this.Unlock()
	this.Pool[task_name] = task
}

func (this *GoPool) AppendFuncTake(task_name string, task_func func(args ...interface{}) (interface{}, error), task_args []interface{}, cyclic bool, step int) {
	task := new_FuncTask(task_name, task_func, task_args, cyclic, step)
	this.Lock()
	defer this.Unlock()
	this.Pool[task_name] = task
}

func (this *GoPool) ModifyOrder(task_name string, new_order string) error {
	this.Lock()
	defer this.Unlock()
	task, ok := this.Pool[task_name]
	if ok {
		task.TaskOrder = new_order
		return nil
	} else {
		return errors.New("don't have this task")
	}
}

func (this *GoPool) ModifyArgs(task_name string, new_args []interface{}) error {
	this.Lock()
	defer this.Unlock()
	task, ok := this.Pool[task_name]
	if ok {
		task.TaskArgs = new_args
		return nil
	} else {
		return errors.New("don't have this task")
	}
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

func (this *GoPool) StartCyclicTask() {
	this.RLock()
	defer this.RUnlock()
	for _, task := range this.Pool {
		if task.Cyclic {
			task.start()
		}
	}
}

func (this *GoPool) StartSingleTask() {
	this.RLock()
	defer this.RUnlock()
	for _, task := range this.Pool {
		if !task.Cyclic {
			task.start()
		}
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

func (this *GoPool) QueryAllTaskState() (status_map map[string]string) {
	this.RLock()
	defer this.RUnlock()
	status_map = make(map[string]string)
	if len(this.Pool) == 0 {
		return
	}
	for name, task := range this.Pool {
		status_map[name] = task.Status
	}
	return
}
