package graceful

import (
	"os"
	"syscall"
)

var restartSignals = []os.Signal{syscall.SIGHUP}
