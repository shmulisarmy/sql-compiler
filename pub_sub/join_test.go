package pubsub

import (
	"sql-compiler/assert"
	"sql-compiler/compiler/rowType"
	"sql-compiler/unwrap"
	"testing"
)

func TestFullOuterJoin(t *testing.T) {
	people := R_Table{
		rowSchema: rowType.RowSchema{
			{Name: "id", Type: rowType.String},
			{Name: "name", Type: rowType.String},
			{Name: "email", Type: rowType.String},
			{Name: "age", Type: rowType.String},
		},
	}
	todos := R_Table{
		rowSchema: rowType.RowSchema{
			{Name: "title", Type: rowType.String},
			{Name: "done", Type: rowType.String},
			{Name: "person_id", Type: rowType.String},
		},
	}

	j := NewFullOuterJoin(&people, &todos, 0, 2)
	j.To_display(unwrap.Some(j.GetRowSchema()))

	people.Add(rowType.RowType{"1", "bob", "email", "22"})
	todos.Add(rowType.RowType{"clean room", "true", "1"})
	people.Add(rowType.RowType{"2", "jan", "email", "22"})
	todos.Add(rowType.RowType{"clean coffee machine", "true", "1"})
	results := []rowType.RowType{}
	for m := range j.Pull {
		results = append(results, m)
	}
	assert.TAssertEq(t, len(results), 2)
	//
	results = []rowType.RowType{}
	people.Add(rowType.RowType{"1", "danny", "email", "22"})
	for m := range j.Pull {
		results = append(results, m)
	}
	assert.TAssertEq(t, len(results), 4)

}
