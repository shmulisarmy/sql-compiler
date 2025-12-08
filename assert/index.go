package assert

import "fmt"

func Assert(condition bool, msg ...string) {
	if len(msg) == 0 {
		msg = []string{"assertion failed"}
	}

	if !condition {
		panic(msg[0])
	}
}

func AssertNot(condition bool, msg ...string) {
	if len(msg) == 0 {
		msg = []string{"assertion failed"}
	}

	if condition {
		panic(msg[0])
	}
}

func AssertEq(a any, b any, msg ...string) {
	if len(msg) == 0 {
		msg = []string{"assertion failed"}
	}
	if a != b {
		panic(fmt.Sprintf("%v != %v: %s", a, b, msg[0]))
	}
}
