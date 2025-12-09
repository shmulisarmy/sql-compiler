package main

import (
	"encoding/json"
	"sql-compiler/compare"
	"sql-compiler/compiler/rowType"
	pubsub "sql-compiler/pub_sub"
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

func TestAddAndRemove(t *testing.T) {

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

	json_string := pubsub.ObserverToJson(&todo_table, row_schema)
	var initial_data_snapshot map[string]map[string]any
	err := json.Unmarshal([]byte(json_string), &initial_data_snapshot)
	if err != nil {
		t.Fatal(err)
	}

	{ //the remove should cancel out the add
		todo_table.Add(rowType.RowType{"clean garage", false, 1})
		todo_table.Remove_where_eq(row_schema, "title", "clean garage")
	}

	json_string = pubsub.ObserverToJson(&todo_table, row_schema)
	var data_snapshot_after_updates map[string]map[string]any
	err = json.Unmarshal([]byte(json_string), &data_snapshot_after_updates)
	if err != nil {
		t.Fatal(err)
	}

	std_message, err := compare.Compare(initial_data_snapshot, data_snapshot_after_updates, "")
	println(std_message)
	if err != nil {
		t.Fatal(err)
	}

}

func TestDoubleUpdateCancelOut(t *testing.T) {

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

	json_string := pubsub.ObserverToJson(&todo_table, row_schema)
	var initial_data_snapshot map[string]map[string]any
	err := json.Unmarshal([]byte(json_string), &initial_data_snapshot)
	if err != nil {
		t.Fatal(err)
	}

	{ //at the end of these 2 updates the data should be as it was previously
		todo_table.Update_where_eq(row_schema, "title", "clean the other room", rowType.RowType{"clean the other room", false, 1})
		todo_table.Update_where_eq(row_schema, "title", "clean the other room", rowType.RowType{"clean the other room", true, 1})
	}

	json_string = pubsub.ObserverToJson(&todo_table, row_schema)
	var data_snapshot_after_updates map[string]map[string]any
	err = json.Unmarshal([]byte(json_string), &data_snapshot_after_updates)
	if err != nil {
		t.Fatal(err)
	}

	std_message, err := compare.Compare(initial_data_snapshot, data_snapshot_after_updates, "")
	println(std_message)
	if err != nil {
		t.Fatal(err)
	}

}
