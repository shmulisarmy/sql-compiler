package assert

func Assert(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}

func AssertNot(condition bool) {
	if condition {
		panic("assertion failed")
	}
}

func AssertEq(a any, b any) {
	if a != b {
		panic("assertion failed")
	}
}
