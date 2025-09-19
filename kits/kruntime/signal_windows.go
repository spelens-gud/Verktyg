package kruntime

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
