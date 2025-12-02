package debugutil

import (
	"fmt"
	"runtime"
)

func Location() string {
	_, filename, line, _ := runtime.Caller(1)
	return filename + ":" + fmt.Sprintf("%d", line)
}
