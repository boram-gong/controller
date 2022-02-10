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
	FAIL    = "Fail"
	TimeOut = "TimeOut"
	SERIOUS = "SeriousErr"
	STOP    = "Stop"
)

type goTask struct {
	sync.RWMutex
	TakeName string
	Args     []interface{}
	Operate  interface{}
	Cyclic   bool
	Step     int
	Status   string
	Note     interface{}
	DownChan chan int
	Result   []interface{}
}

func newGoTask(taskName string, fun interface{}, args []interface{}, cyclic bool, step int, note interface{}) (task *goTask) {
	task = new(goTask)
	task.TakeName = taskName
	task.Operate = fun
	task.Args = args
	task.Cyclic = cyclic
	task.Step = step
	task.Status = INIT
	task.DownChan = make(chan int, 1)
	task.Note = note
	return
}

func (gt *goTask) start(ch chan []interface{}) {
	if gt.Status == START {
		LogChan <- fmt.Sprintf("WARN: '%v' task is running, don't to need run again", gt.TakeName)
		return
	}
	gt.Status = START
	go gt.work(ch)
}

func (gt *goTask) work(ch chan []interface{}) {
	if gt.run(ch) {
		if gt.Status != SERIOUS && gt.Status != FAIL && !gt.Cyclic {
			gt.Status = SUCCESS + "_" + time.Now().Format("2006/01/02/15:04")
		}
	}
	if !gt.Cyclic {
		return
	}
	for {
		select {
		case <-time.Tick(time.Duration(gt.Step) * time.Second):
			if !gt.run(ch) {
				return
			}
		case <-gt.DownChan:
			LogChan <- fmt.Sprintf("INFO: '%v' task manual stop", gt.TakeName)
			return
		}
	}
}

func (gt *goTask) stop() {
	if !gt.Cyclic {
		return
	}
	if gt.Status == STOP {
		return
	}
	gt.DownChan <- 1
	gt.Lock()
	gt.Status = STOP
	gt.Unlock()
}

func (gt *goTask) run(ch chan []interface{}) bool {
	var (
		result []interface{}
	)
	result = gt.funGenerator()
	if gt.Status == SERIOUS || gt.Status == FAIL {
		return false
	}
	if len(result) != 0 {
		ch <- result
		gt.Result = result[1:]
	}
	return true
}

func (gt *goTask) funGenerator() []interface{} {
	defer func(g *goTask) {
		serious := recover()
		if serious != nil {
			LogChan <- fmt.Sprintf("panic: %s task fail: %v", g.TakeName, serious)
			g.Status = SERIOUS
			if g.Cyclic {
				g.DownChan <- 1
			}
		}
	}(gt)
	var (
		rFun              reflect.Value
		args, returnValue []reflect.Value
	)
	rFun = reflect.ValueOf(gt.Operate)
	if rFun.Kind() != reflect.Func {
		LogChan <- fmt.Sprintf("ERROR: %s no task", gt.TakeName)
		gt.Status = FAIL
		return nil
	}
	for _, a := range gt.Args {
		args = append(args, reflect.ValueOf(a))
	}
	returnValue = rFun.Call(args)
	if len(returnValue) != 0 {
		l := make([]interface{}, len(returnValue)+1)
		l[0] = gt.TakeName
		for i, r := range returnValue {
			l[i+1] = r.Interface()
		}
		return l
	}
	return nil
}
