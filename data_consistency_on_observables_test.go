package main

import (
	"encoding/json"
	"sql-compiler/compare"
	pubsub "sql-compiler/pub_sub"
	"sql-compiler/rowType"
	"testing"
)

func Test_Adds(t *testing.T) {

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

	expected := map[string]map[string]any{
		"clean the room": {
			"title":     "clean the room",
			"completed": true,
			"person_id": 1,
		},
		"clean the other room": {
			"title":     "clean the other room",
			"completed": true,
			"person_id": 1,
		},
	}

	json_string := pubsub.ObserverToJson(&todo_table, row_schema)
	var actual map[string]map[string]any
	err := json.Unmarshal([]byte(json_string), &actual)
	if err != nil {
		t.Fatal(err)
	}

	std_message, err := compare.Compare(expected, actual, "")
	println(std_message)
	if err != nil {
		t.Fatal(err)
	}

}
