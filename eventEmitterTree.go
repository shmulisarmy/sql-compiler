// right now, the only way to test the functionality in this file is to run an
// integration test that includes the front end where @run_tests.tsx the front end has two
// objects that are a reflection of how a query view that gets updated via an
// event tree and make sure that the two of them are in sync. However, it would
// be a nice to have a unit test for this too, as it will then be able to be ran
// more often and quicker.

package main

import (
	"sql-compiler/compiler/rowType"
	pubsub "sql-compiler/pub_sub"
)

const path_separator = "/"

type SyncType string

const (
	SyncTypeUpdate  = "update"
	SyncTypeAdd     = "add"
	SyncTypeRemove  = "remove"
	LoadInitialData = "load"
)

type SyncMessage struct {
	Type      SyncType
	Data      string
	Path      string
	Timestamp int64
}

type eventEmitterTree struct {
	on_message func(SyncMessage)
}

func (receiver *eventEmitterTree) syncFromObservable(obs *pubsub.Mapper, path string) {
	// switch obs := obs.(type) {
	// case *pubsub.Mapper:
	// 	for row := range obs.Pull {
	// 		pubsub.RowTypeToJson(&row, obs.RowSchema.Unwrap())
	// 	}
	// default:
	// 	panic("expected mapper")
	// }
	obs.Add_sub(&pubsub.CustomSubscriber{
		OnAddFunc: func(item rowType.RowType) {
			primary_key := item[0].(string)
			receiver.on_message(SyncMessage{Type: SyncTypeAdd, Data: pubsub.RowTypeToJson(&item, obs.RowSchema.Unwrap()), Path: path + path_separator + primary_key})
			receiver.syncFromObservable_row(item, path+path_separator+primary_key, obs.RowSchema.Unwrap())
		},
		OnRemoveFunc: func(item rowType.RowType) {
			primary_key := item[0].(string)
			receiver.on_message(SyncMessage{Type: SyncTypeRemove, Data: pubsub.RowTypeToJson(&item, obs.RowSchema.Unwrap()), Path: path + path_separator + primary_key})
		},
		OnUpdateFunc: func(oldItem, newItem rowType.RowType) {
			primary_key := oldItem[0].(string)
			receiver.on_message(SyncMessage{Type: SyncTypeUpdate, Data: pubsub.RowTypeToJson(&newItem, obs.RowSchema.Unwrap()), Path: path + path_separator + primary_key})
		},
	})
	for row := range obs.Pull {
		receiver.syncFromObservable_row(row, path, obs.RowSchema.Unwrap())
	}

}

func (receiver *eventEmitterTree) syncFromObservable_row(row rowType.RowType, path string, row_schema rowType.RowSchema) {
	for i, col := range row {
		switch col := col.(type) {
		case string, int, bool:
		case pubsub.ObservableI:
			switch col := col.(type) {
			case *pubsub.Mapper:
				receiver.syncFromObservable(col, path+path_separator+row_schema[i].Name)
			default:
				panic("should be mapper")
			}
		default:
			panic("unhandled")
		}
	}
}
