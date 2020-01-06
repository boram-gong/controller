package controller

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var system = runtime.GOOS

func cmdWork(shell string, sec int) (string, error) {
	var (
		done     = make(chan error, 1)
		content  = make(chan []byte, 1)
		time_out = time.Duration(time.Duration(sec) * time.Second)
	)
	if system == "linux" {
		go func() {
			cmd := exec.Command("/bin/bash", "-c", shell)
			output, err := cmd.CombinedOutput()
			if err != nil {
				done <- err
				return
			}
			content <- output
		}()
	} else if system == "windows" {
		go func() {
			cmd := exec.Command("CMD", "/C", shell)
			output, err := cmd.CombinedOutput()
			if err != nil {
				done <- err
				return
			}
			content <- output
		}()
	} else {
		done <- errors.New("this system <" + system + "> is not supported")
	}
	if sec <= 0 {
		select {
		case err := <-done:
			return "", err

		case out := <-content:
			return string(out), nil
		}
	} else {
		select {
		case <-time.After(time_out):
			return "", errors.New(TimeOut)

		case err := <-done:
			return "", err

		case out := <-content:
			return string(out), nil
		}
	}
}

func outStringDeal(str string) string {
	return strings.Split(str, "\n")[0]
}

func ArgsMaker(arg ...interface{}) (args []interface{}) {
	args = append(args, arg...)
	return
}
