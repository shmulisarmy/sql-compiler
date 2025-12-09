package main

import (
	"encoding/json"
	"sql-compiler/assert"
	"sql-compiler/compare"
	"sql-compiler/compiler/rowType"
	pubsub "sql-compiler/pub_sub"
	"testing"
)

func TestThatMapperIsAlwaysInSync(t *testing.T) {
	//a mapper should have the same amount of rows as the the table that its pointed to

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
	todo_table := pubsub.New_R_Table(row_schema)

	todo_table.Add(rowType.RowType{
		"clean the room", true, 1,
	})
	todo_table.Add(rowType.RowType{
		"clean the other room", true, 1,
	})
	todo_table.Add(rowType.RowType{
		"take out the trash", false, 1,
	})
	mapper := todo_table.Map_on(func(rt rowType.RowType) rowType.RowType {
		return rowType.RowType{rt[0].(string) + "!"}
	})
	json_string := pubsub.ObserverToJson(mapper, row_schema)
	var actual map[string]map[string]any
	err := json.Unmarshal([]byte(json_string), &actual)
	if err != nil {
		t.Fatal(err)
	}

	assert.AssertEq(len(actual), 3)
	todo_table.Add(rowType.RowType{
		"finish making food", true, 1,
	})
	json_string = pubsub.ObserverToJson(mapper, row_schema)
	err = json.Unmarshal([]byte(json_string), &actual)
	if err != nil {
		t.Fatal(err)
	}
	assert.AssertEq(len(actual), 4)

}

func TestThatMapperAgainstExpected(t *testing.T) {
	//a mapper should have the same amount of rows as the the table that its pointed to

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
	todo_table := pubsub.New_R_Table(row_schema)

	todo_table.Add(rowType.RowType{
		"clean the room", true, 1,
	})
	todo_table.Add(rowType.RowType{
		"clean the other room", true, 1,
	})
	todo_table.Add(rowType.RowType{
		"take out the trash", false, 1,
	})
	mapper := todo_table.Map_on(func(rt rowType.RowType) rowType.RowType {
		return rowType.RowType{rt[0].(string) + "!"}
	})
	json_string := pubsub.ObserverToJson(mapper, row_schema)
	var actual map[string]map[string]any
	err := json.Unmarshal([]byte(json_string), &actual)
	if err != nil {
		t.Fatal(err)
	}

	var expected = map[string]map[string]any{
		"clean the room!": {
			"title": "clean the room!",
		},
		"clean the other room!": {
			"title": "clean the other room!",
		},
		"take out the trash!": {
			"title": "take out the trash!",
		},
	}
	output, err := compare.Compare(expected, actual, "")
	println(output)
	if err != nil {
		t.Fatal(err)
	}

}
