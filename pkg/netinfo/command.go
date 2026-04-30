package netinfo

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type runCommandFunc func(name string, args ...string) ([]byte, error)
type runCommandWithTimeoutFunc func(timeout time.Duration, name string, args ...string) ([]byte, error)

var commandTimeout = 3 * time.Second
var traceCommandTimeout = 45 * time.Second

var commandRunnerWithTimeout runCommandWithTimeoutFunc = func(timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := exec.CommandContext(ctx, name, args...).Output()
	if ctx.Err() == context.DeadlineExceeded {
		return output, fmt.Errorf("%s timed out after %s", name, timeout)
	}
	return output, err
}

var commandRunner runCommandFunc = func(name string, args ...string) ([]byte, error) {
	return commandRunnerWithTimeout(commandTimeout, name, args...)
}
