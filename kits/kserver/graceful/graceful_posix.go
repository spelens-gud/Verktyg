//go:build !windows
// +build !windows

package graceful

import (
	"os"
	"syscall"
)

var RestartSignals = []os.Signal{syscall.SIGUSR1}
