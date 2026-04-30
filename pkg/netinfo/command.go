package netinfo

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type runCommandFunc func(name string, args ...string) ([]byte, error)

var commandTimeout = 3 * time.Second

var commandRunner runCommandFunc = func(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	output, err := exec.CommandContext(ctx, name, args...).Output()
	if ctx.Err() == context.DeadlineExceeded {
		return output, fmt.Errorf("%s timed out after %s", name, commandTimeout)
	}
	return output, err
}
