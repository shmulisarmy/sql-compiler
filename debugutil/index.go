package debugutil

import (
	"fmt"
	"runtime"
)

func Location() string {
	_, filename, line, _ := runtime.Caller(1)
	return filename + ":" + fmt.Sprintf("%d", line)
}

func Print(item_to_display any, name string) {
	_, filename, line, _ := runtime.Caller(1)
	location := filename + ":" + fmt.Sprintf("%d", line)
	fmt.Print(location + ": ")
	fmt.Printf(name, "=", item_to_display, "\n")
}
