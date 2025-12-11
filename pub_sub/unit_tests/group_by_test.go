package main

import (
	"sql-compiler/assert"
	"sql-compiler/compiler/rowType"
	compiler_runtime "sql-compiler/compiler/runtime"
	"sql-compiler/db_tables"
	"sql-compiler/display"
	event_emitter_tree "sql-compiler/eventEmitterTree"
	"sql-compiler/local_live_db"
	"sql-compiler/unwrap"
	"testing"
)

//since this structure is one that only exists if eventEmitterTree responds to Group in the correct way this is really testing EventEmitterTree

func TestGroupBy(t *testing.T) {
	sql := "select id, name, age from person where age >= 0 group by age "
	obs := compiler_runtime.Query_to_observer(sql)

	obs.To_display(unwrap.None[rowType.RowSchema]())

	// @ANTI_SOLID_PATTERN

	live_db := local_live_db.LocalLiveDB{
		Data: make(map[string]any),
	}
	event_emitter := event_emitter_tree.EventEmitterTree{
		On_message: func(message event_emitter_tree.SyncMessage) {
			live_db.HandleUpdate(message)
		},
	}
	event_emitter.SyncFromObservable(obs, "")
	db_tables.Tables.Get("person").Insert(rowType.RowType{"shmuli", "email@gmail.com", 22, "state", db_tables.Tables.Get("person").Next_row_id()})
	db_tables.Tables.Get("person").Insert(rowType.RowType{"ajay", "ajay@gmail.com", 30, "state", db_tables.Tables.Get("person").Next_row_id()})
	db_tables.Tables.Get("person").Insert(rowType.RowType{"natalie", "natalie@gmail.com", 22, "state", db_tables.Tables.Get("person").Next_row_id()})
	db_tables.Tables.Get("person").Insert(rowType.RowType{"ellen", "ellen@gmail.com", 30, "state", db_tables.Tables.Get("person").Next_row_id()})
	db_tables.Tables.Get("person").Insert(rowType.RowType{"fred", "fred@gmail.com", 44, "state", db_tables.Tables.Get("person").Next_row_id()})
	// Check that the top-level age groups exist in the local_live_db after inserts

	_, has22 := live_db.Data["22"]
	_, has30 := live_db.Data["30"]
	_, has44 := live_db.Data["44"]
	assert.TAssert(t, has22, "expected key 22 to exist")
	assert.TAssert(t, has30, "expected key 30 to exist")
	assert.TAssert(t, has44, "expected key 44 to exist")

	display.DisplayStruct(live_db.Data)
}
