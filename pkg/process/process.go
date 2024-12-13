package process

import (
	"os/exec"
)

// Execute the command
func Execute(command string, arg ...string) (string, error) {
	o, err := exec.Command(command, arg...).CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(o[:]), nil
}
