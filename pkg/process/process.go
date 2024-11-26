package process

import (
	"os/exec"
)

// Execute the command
func Execute(command string, arg ...string) (string, error) {
	o, err := exec.Command(command, arg...).CombinedOutput()
	if err != nil && len(o) != 0 {
		return "", err
	} else if err != nil && len(o) == 0 {
		return "", nil
	}
	return string(o[:]), nil
}
