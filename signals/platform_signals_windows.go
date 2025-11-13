//go:build windows

package signals

import "os"

var platformSpecificSignals = map[string]os.Signal{}
