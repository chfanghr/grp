package grp

import (
	"runtime"
)

var processNumber int

func defaultConfig() {
	processNumber = runtime.NumCPU()
}
