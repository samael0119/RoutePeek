package netinfo

import "os/exec"

type runCommandFunc func(name string, args ...string) ([]byte, error)

var commandRunner runCommandFunc = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}
