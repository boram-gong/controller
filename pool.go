package controller

import (
	"sync"
)

type TaskPool struct {
	Pool        sync.Map
	OutDataChan chan []interface{}
}

func NewPool() (pool *TaskPool) {
	pool = new(TaskPool)
	pool.OutDataChan = make(chan []interface{}, 1000)
	return
}

func (p *TaskPool) ResetMessagePool(length int) {
	p.OutDataChan = nil
	p.OutDataChan = make(chan []interface{}, length)
}

func (p *TaskPool) AppendFuncTake(taskName string, fun interface{}, args []interface{}, note interface{}) {
	task := newGoTask(taskName, fun, args, false, 0, note)
	p.Pool.Store(taskName, task)
}

func (p *TaskPool) AppendCyclicFuncTake(taskName string, fun interface{}, args []interface{}, cyclic bool, step int, note interface{}) {
	task := newGoTask(taskName, fun, args, cyclic, step, note)
	p.Pool.Store(taskName, task)
}

func (p *TaskPool) StartAllTake() {
	p.Pool.Range(func(k, v interface{}) bool {
		v.(*goTask).start(p.OutDataChan)
		return true
	})
}

func (p *TaskPool) StartOneTask(taskName string) bool {
	task, ok := p.Pool.Load(taskName)
	if ok {
		task.(*goTask).start(p.OutDataChan)
		return true
	} else {
		return false
	}
}

func (p *TaskPool) StopAllTake() {
	p.Pool.Range(func(k, v interface{}) bool {
		v.(*goTask).stop()
		return true
	})
}

func (p *TaskPool) StopOneTask(taskName string) {
	task, ok := p.Pool.Load(taskName)
	if ok {
		task.(*goTask).stop()
	}
}

func (p *TaskPool) DeleteTake(taskName string) {
	p.StartOneTask(taskName)
	p.Pool.Delete(taskName)
}

func (p *TaskPool) DeleteAllTake() {
	p.StopAllTake()
	p.Pool = sync.Map{}
}

func (p *TaskPool) QueryTaskMessage(taskName string) TaskMsg {
	task, ok := p.Pool.Load(taskName)
	if ok {
		return task.(*goTask).TaskMsg
	} else {
		return TaskMsg{}
	}
}

func ArgsMaker(arg ...interface{}) (args []interface{}) {
	args = append(args, arg...)
	return
}
