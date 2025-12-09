package compiler_runtime

import (
	"encoding/json"
	"sql-compiler/compare"
	"sql-compiler/compiler/rowType"
	"sql-compiler/db_tables"
	pubsub "sql-compiler/pub_sub"
	"testing"
)

func TestObservedLists(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic: %v", r)
		}
	}()
	src := `SELECT person.name, person.email, person.id FROM person `
	people := db_tables.Tables.Get("person")
	id := len(people.R_Table.Rows)
	people.Insert(rowType.RowType{"example-name", "example-email", 23, "state", id})
	people.Insert(rowType.RowType{"example-name-2", "example-email-2", 23, "state-2", id})

	obs := Query_to_observer(src)

	expected := map[string]any{
		"example-name": map[string]any{
			"name":  "example-name",
			"email": "example-email",
			"id":    id,
		},
		"example-name-2": map[string]any{
			"name":  "example-name-2",
			"email": "example-email-2",
			"id":    id,
		},
	}

	json_string := pubsub.ObserverToJson(obs, obs.RowSchema.Unwrap())
	var actual_ast map[string]any
	json.Unmarshal([]byte(json_string), &actual_ast)

	std_message, err := compare.Compare(expected, actual_ast, "")
	if err != nil {
		t.Log(std_message)
		t.Fatal(err)
	}
	t.Log(std_message)
}
