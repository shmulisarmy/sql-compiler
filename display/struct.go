package display

// display_struct.go

import (
	"fmt"
	"reflect"
	"sql-compiler/colors"
	"strings"
)

const (
	red    = "\033[31m"
	green  = "\033[32m"
	blue   = "\033[34m"
	yellow = "\033[33m"
	reset  = "\033[0m"
)

func color(s, c string) string {
	return c + s + reset
}

func DisplayStruct(v interface{}) {
	out := displayValue(reflect.ValueOf(v), 0)
	fmt.Println(colors.Blue(out))
}

func displayValue(v reflect.Value, indent int) string {
	ind := strings.Repeat("  ", indent)

	// pointer
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ind + color("nil", red)
		}
		return displayValue(v.Elem(), indent)
	}

	switch v.Kind() {

	case reflect.Struct:
		out := ind + color(v.Type().String(), yellow) + " {\n"
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			// Skip unexported fields
			if !f.IsExported() {
				continue
			}
			fv := v.Field(i)
			out += fmt.Sprintf("%s  %s: %s\n",
				ind,
				color(f.Name, green),
				displayValue(fv, indent+1),
			)
		}
		return out + ind + "}"

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return ind + color("[]", red)
		}
		out := ind + "[\n"
		for i := 0; i < v.Len(); i++ {
			out += displayValue(v.Index(i), indent+1) + "\n"
		}
		return out + ind + "]"

	case reflect.Map:
		if v.Len() == 0 {
			return ind + color("{}", red)
		}
		out := ind + "{\n"
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key()
			val := iter.Value()
			out += fmt.Sprintf(
				"%s  %s: %s\n",
				ind,
				color(strings.TrimSpace(displayValue(k, 0)), blue),
				displayValue(val, indent+1),
			)
		}
		return out + ind + "}"

	case reflect.String:
		return ind + color(fmt.Sprintf("%q", v.String()), green)

	case reflect.Interface:
		if v.IsNil() {
			return ind + color("nil", red)
		}
		return displayValue(v.Elem(), indent)

	default:
		return ind + color(fmt.Sprintf("%v", v.Interface()), blue)
	}
}
