package controller

import (
	"context"
	"log"
	"reflect"
	"time"
)

const (
	INIT     = "Init"
	RUN      = "Running"
	COMPLETE = "Complete"
	FAIL     = "Fail"
	PANIC    = "Panic"
	STOP     = "Stop"
)

type goTask struct {
	ctx     context.Context
	clear   context.CancelFunc
	Operate interface{} `json:"-"`
	TaskMsg
}

type TaskMsg struct {
	Args     []interface{} `json:"args"`
	Cyclic   bool          `json:"cyclic"`
	Step     int           `json:"step"`
	Status   string        `json:"status"`
	LastTime time.Time     `json:"lastTime"`
	Note     interface{}   `json:"note"`
	Result   []interface{} `json:"result"`
}

func newGoTask(taskName string, fun interface{}, args []interface{}, cyclic bool, step int, note interface{}) (task *goTask) {
	task = new(goTask)
	task.ctx, task.clear = context.WithCancel(context.WithValue(context.Background(), "name", taskName))
	task.Operate = fun
	task.Args = args
	task.Cyclic = cyclic
	task.Step = step
	task.Status = INIT
	task.Note = note
	return
}

func (gt *goTask) start(ch chan []interface{}) {
	if gt.Status == RUN {
		return
	}
	gt.Status = RUN
	go gt.work(ch)
}

func (gt *goTask) work(ch chan []interface{}) {
	if gt.run(ch) {
		if gt.Status != PANIC && gt.Status != FAIL && !gt.Cyclic {
			gt.Status = COMPLETE
		}
	}
	if gt.Cyclic {
		for _ = range time.Tick(time.Duration(gt.Step) * time.Second) {
			select {
			case <-gt.ctx.Done():
				return
			default:
				if !gt.run(ch) {
					return
				}
			}
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
	gt.clear()
	gt.Status = STOP
}

func (gt *goTask) run(ch chan []interface{}) bool {
	var (
		result []interface{}
	)
	result = gt.funGenerator()
	gt.LastTime = time.Now()
	if gt.Status == PANIC || gt.Status == FAIL {
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
			log.Printf("panic: %s task fail: %v \n", gt.ctx.Value("name"), serious)
			g.Status = PANIC
			if g.Cyclic {
				g.clear()
			}
		}
	}(gt)
	var (
		rFun              reflect.Value
		args, returnValue []reflect.Value
	)
	rFun = reflect.ValueOf(gt.Operate)
	if rFun.Kind() != reflect.Func {
		gt.Status = FAIL
		return nil
	}
	for _, a := range gt.Args {
		args = append(args, reflect.ValueOf(a))
	}
	returnValue = rFun.Call(args)
	if len(returnValue) != 0 {
		l := make([]interface{}, len(returnValue)+1)
		l[0] = gt.ctx.Value("name")
		for i, r := range returnValue {
			l[i+1] = r.Interface()
		}
		return l
	}
	return nil
}
