package controller

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

var system = runtime.GOOS

func cmdWork(shell string) (string, error) {
	var (
		cmd    *exec.Cmd
		output []byte
		err    error
	)

	if system == "linux" {
		cmd = exec.Command("/bin/bash", "-c", shell)
		if output, err = cmd.CombinedOutput(); err != nil {
			return "", err
		}
	} else if system == "windows" {
		cmd = exec.Command("CMD", "/C", shell)
		if output, err = cmd.CombinedOutput(); err != nil {
			return "", err
		}
	} else {
		return "", errors.New("this system <"+system+"> is not supported")
	}


	return string(output), nil
}

func outStringDeal(str string) string {
	return strings.Split(str, "\n")[0]
}
