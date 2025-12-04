package pubsub

import (
	"fmt"
	. "sql-compiler/rowType"
)

func RowTypeToJson(row *RowType, row_schema RowSchema) string {
	res := "{"
	for i, col := range *row {
		res += "\"" + row_schema[i].Name + "\":"
		switch row_schema[i].Type {
		case String:
			res += fmt.Sprintf("\\\"%s\\\"", col.(string))
		case Int:
			res += fmt.Sprintf("%d", col.(int))
		case Bool:
			res += fmt.Sprintf("%t", col.(bool))
		default:
			res += "["
			for row := range col.(ObservableI).Pull {
				res += RowTypeToJson(&row, row_schema) + ","
			}
			res += "]"
		}
		if i != len(*row)-1 {
			res += ","
		}
	}
	return res
}
