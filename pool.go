package controller

import (
	"sync"
)

var LogChan = make(chan string, 1000)

type TaskPool struct {
	sync.RWMutex
	Pool        map[string]*goTask
	OutDataChan chan []interface{}
}

func NewPool() (pool *TaskPool) {
	pool = new(TaskPool)
	pool.Pool = make(map[string]*goTask)
	pool.OutDataChan = make(chan []interface{}, 1000)
	return
}

func (p *TaskPool) ResetMessagePool(length int) {
	p.OutDataChan = nil
	p.OutDataChan = make(chan []interface{}, length)
}

func (p *TaskPool) AppendFuncTake(taskName string, fun interface{}, args []interface{}, note interface{}) {
	task := newGoTask(taskName, fun, args, false, 0, note)
	p.Lock()
	defer p.Unlock()
	p.Pool[taskName] = task
}

func (p *TaskPool) AppendCyclicFuncTake(taskName string, fun interface{}, args []interface{}, cyclic bool, step int, note interface{}) {
	task := newGoTask(taskName, fun, args, cyclic, step, note)
	p.Lock()
	defer p.Unlock()
	p.Pool[taskName] = task
}

func (p *TaskPool) DeleteTake(taskName string) {
	p.Lock()
	defer p.Unlock()
	task, ok := p.Pool[taskName]
	if ok {
		task.stop()
	}
	delete(p.Pool, taskName)
}

func (p *TaskPool) DeleteAllTake() {
	p.Lock()
	defer p.Unlock()
	for _, task := range p.Pool {
		task.stop()
	}
	p.Pool = nil
	p.Pool = make(map[string]*goTask)
}

func (p *TaskPool) StartAllTake() {
	p.RLock()
	defer p.RUnlock()
	for _, task := range p.Pool {
		task.start(p.OutDataChan)
	}
}

func (p *TaskPool) StartOneTask(taskName string) bool {
	p.RLock()
	defer p.RUnlock()
	task, ok := p.Pool[taskName]
	if !ok {
		return false
	} else {
		task.start(p.OutDataChan)
		return true
	}
}

func (p *TaskPool) StartAllCyclicTask() {
	p.RLock()
	defer p.RUnlock()
	for _, task := range p.Pool {
		if task.Cyclic {
			task.start(p.OutDataChan)
		}
	}
}

func (p *TaskPool) StartAllNotCyclicTask() {
	p.RLock()
	defer p.RUnlock()
	for _, task := range p.Pool {
		if !task.Cyclic {
			task.start(p.OutDataChan)
		}
	}
}

func (p *TaskPool) StopAllTake() {
	p.RLock()
	defer p.RUnlock()
	for _, task := range p.Pool {
		task.stop()
	}
}

func (p *TaskPool) StopOneTask(taskName string) bool {
	p.RLock()
	defer p.RUnlock()
	task, ok := p.Pool[taskName]
	if !ok {
		return false
	} else {
		task.stop()
		return true
	}
}

func (p *TaskPool) QueryTaskProperty(taskName string) map[string]interface{} {
	p.RLock()
	defer p.RUnlock()
	task, ok := p.Pool[taskName]
	if ok {
		m := make(map[string]interface{})
		m["TaskName"] = task.TakeName
		m["Status"] = task.Status
		m["Operate"] = task.Operate
		m["TaskArgs"] = task.Args
		m["Cyclic"] = task.Cyclic
		m["Step"] = task.Step
		m["Result"] = task.Result
		return m
	} else {
		return nil
	}
}

func (p *TaskPool) QueryAllTaskState() (status map[string]string) {
	p.RLock()
	defer p.RUnlock()
	status = make(map[string]string)
	if len(p.Pool) == 0 {
		return
	}
	for name, task := range p.Pool {
		status[name] = task.Status
	}
	return
}

func (p *TaskPool) QueryTaskNote(taskName string) (interface{}, bool) {
	p.RLock()
	defer p.RUnlock()
	task, ok := p.Pool[taskName]
	return task.Note, ok
}

func ArgsMaker(arg ...interface{}) (args []interface{}) {
	args = append(args, arg...)
	return
}
