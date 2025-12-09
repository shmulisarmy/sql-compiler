package assert

import (
	"testing"
)

func TAssert(t *testing.T, condition bool, msg ...string) {
	if len(msg) == 0 {
		msg = []string{"assertion failed"}
	}

	if !condition {
		t.Fatal(msg[0])
	}
}

func TAssertNot(t *testing.T, condition bool, msg ...string) {
	if len(msg) == 0 {
		msg = []string{"assertion failed"}
	}

	if condition {
		t.Fatal(msg[0])
	}
}

func TAssertEq(t *testing.T, a, b any, msg ...string) {
	if len(msg) == 0 {
		msg = []string{"assertion failed"}
	}
	if a != b {
		t.Fatalf("%v != %v: %s", a, b, msg[0])
	}
}
