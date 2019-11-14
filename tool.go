package controller

import "os/exec"

func CmdWork(shell string) (string, error) {
	var (
		cmd    *exec.Cmd
		output []byte
		err    error
	)
	cmd = exec.Command("/bin/bash", "-c", shell)

	if output, err = cmd.CombinedOutput(); err != nil {
		return "", err
	}

	return string(output), nil
}
