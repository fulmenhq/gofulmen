//go:build !windows

package signals

import (
	"os"
	"syscall"
)

var platformSpecificSignals = map[string]os.Signal{
	"SIGUSR1": syscall.SIGUSR1,
	"SIGUSR2": syscall.SIGUSR2,
}
