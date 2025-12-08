package main

import (
	"encoding/json"
	"sql-compiler/assert"
	pubsub "sql-compiler/pub_sub"
	"sql-compiler/rowType"
	"testing"
)

func TestThatFilterIsAlwaysInSync(t *testing.T) {

	row_schema := rowType.RowSchema{
		rowType.ColInfo{
			Type: rowType.String,
			Name: "title",
		},
		rowType.ColInfo{
			Type: rowType.Bool,
			Name: "completed",
		},
		rowType.ColInfo{
			Type: rowType.Int,
			Name: "person_id",
		},
	}
	todo_table := pubsub.New_R_Table()

	todo_table.Add(rowType.RowType{
		"clean the room", true, 1,
	})
	todo_table.Add(rowType.RowType{
		"clean the other room", true, 1,
	})
	todo_table.Add(rowType.RowType{
		"take out the trash", false, 1,
	})
	filter := todo_table.Filter_on(func(rt rowType.RowType) bool {
		return rt[1].(bool) == true
	})
	json_string := pubsub.ObserverToJson(filter, row_schema)
	var actual map[string]map[string]any
	err := json.Unmarshal([]byte(json_string), &actual)
	if err != nil {
		t.Fatal(err)
	}

	assert.AssertEq(len(actual), 2)
	todo_table.Add(rowType.RowType{
		"finish making food", true, 1,
	})
	json_string = pubsub.ObserverToJson(filter, row_schema)
	err = json.Unmarshal([]byte(json_string), &actual)
	if err != nil {
		t.Fatal(err)
	}
	assert.AssertEq(len(actual), 3)

}
