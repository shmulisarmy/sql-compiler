package utils

import (
	"fmt"
	"strings"
)

func String_or_num_to_string(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case int:
		return fmt.Sprintf("%d", value)
	default:
		panic("only string and int are supported")
	}
}
func Capitalize(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func CompareSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
