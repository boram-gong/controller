package controller

import (
	"github.com/pkg/errors"
	"sync"
)

var LogChan = make(chan string, 1000)

// 协程池结构体
type GoPool struct {
	sync.RWMutex
	Pool        map[string]*goTask
	OutDataChan chan []interface{}
}

// new一个协程池
func NewPool() (pool *GoPool) {
	pool = new(GoPool)
	pool.Pool = make(map[string]*goTask)
	pool.OutDataChan = make(chan []interface{}, 1000)
	return
}

// 修改协程池的数据管道容量
func (this *GoPool) ResetMessagePool(length int) {
	this.OutDataChan = nil
	this.OutDataChan = make(chan []interface{}, length)
}

// 添加一个指令型任务
func (this *GoPool) AppendOrderTake(task_name string, task_order string, cyclic bool, step int) {
	task := new_OrderTask(task_name, task_order, cyclic, step)
	this.Lock()
	defer this.Unlock()
	this.Pool[task_name] = task
}

// 添加一个函数型任务
func (this *GoPool) AppendFuncTake(task_name string, task_func interface{}, task_args []interface{}, cyclic bool, step int) {
	task := new_FuncTask(task_name, task_func, task_args, cyclic, step)
	this.Lock()
	defer this.Unlock()
	this.Pool[task_name] = task
}

// 修改命令
func (this *GoPool) ModifyOrder(task_name string, new_order string) error {
	this.Lock()
	defer this.Unlock()
	task, ok := this.Pool[task_name]
	if ok {
		task.TaskOrder = new_order
		this.Pool[task_name] = task
		return nil
	} else {
		return errors.New("don't have this task")
	}
}

// 修改参数
func (this *GoPool) ModifyArgs(task_name string, new_args []interface{}) error {
	this.Lock()
	defer this.Unlock()
	task, ok := this.Pool[task_name]
	if ok {
		task.TaskArgs = new_args
		this.Pool[task_name] = task
		return nil
	} else {
		return errors.New("don't have this task")
	}
}

// 删除指定任务
func (this *GoPool) DeleteTake(task_name string) {
	this.Lock()
	defer this.Unlock()
	task, ok := this.Pool[task_name]
	if ok {
		task.stop()
	}
	delete(this.Pool, task_name)
}

// 删除全部任务
func (this *GoPool) DeleteAllTake() {
	this.Lock()
	defer this.Unlock()
	for _, task := range this.Pool {
		task.stop()
	}
	this.Pool = nil
	this.Pool = make(map[string]*goTask)
}

// 启动全部任务
func (this *GoPool) StartAllTake() {
	this.RLock()
	defer this.RUnlock()
	for _, task := range this.Pool {
		task.start(this.OutDataChan)
	}
}

// 启动指定任务
func (this *GoPool) StartOneTask(task_name string) bool {
	this.RLock()
	defer this.RUnlock()
	task, ok := this.Pool[task_name]
	if !ok {
		return false
	} else {
		task.start(this.OutDataChan)
		return true
	}
}

// 启动所有循环任务
func (this *GoPool) StartCyclicTask() {
	this.RLock()
	defer this.RUnlock()
	for _, task := range this.Pool {
		if task.Cyclic {
			task.start(this.OutDataChan)
		}
	}
}

// 启动所有单次任务
func (this *GoPool) StartSingleTask() {
	this.RLock()
	defer this.RUnlock()
	for _, task := range this.Pool {
		if !task.Cyclic {
			task.start(this.OutDataChan)
		}
	}
}

// 停止所有任务
func (this *GoPool) StopAllTake() {
	this.RLock()
	defer this.RUnlock()
	for _, task := range this.Pool {
		task.stop()
	}
}

// 停止指定任务
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

// 查询指定任务的任务状态
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

// 查询所有任务状态
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
