package pubsub

import (
	"fmt"
	"sql-compiler/assert"
	. "sql-compiler/rowType"
	"strings"
)

func RowTypeToJson(row *RowType, row_schema RowSchema) string {
	res := "{"
	for i, col := range *row {
		res += "\"" + row_schema[i].Name + "\":"
		switch row_schema[i].Type {
		case String:
			res += fmt.Sprintf("\"%s\"", col.(string))
		case Int:
			res += fmt.Sprintf("%d", col.(int))
		case Bool:
			res += fmt.Sprintf("%t", col.(bool))
		default:
			childs_row_schema := NestedSelectsRowSchema[row_schema[i].Type]
			res += ObserverToJson(col.(ObservableI), childs_row_schema)
		}
		if i != len(*row)-1 {
			res += ","
		}
	}
	res += "}"
	return res
}

func ObserverToJson(col ObservableI, row_schema RowSchema) string {
	res := "{"
	has_at_least_one := false
	for row := range col.Pull {
		primary_key := row[0].(string)
		res += "\"" + primary_key + "\":"
		res += RowTypeToJson(&row, row_schema) + ","
		has_at_least_one = true
	}
	if !has_at_least_one {
		return "{}"
	}
	assert.Assert(res[len(res)-1] == ',')
	res = strings.TrimSuffix(res, ",")
	return res + "}"
}
