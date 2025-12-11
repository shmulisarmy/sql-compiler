// the point of this structure is to take an observable (that has subqueries) and subscribe to it
// in a tree like manner where as it goes down into a subquery it builds up a path (that gets put into the message), and whatever is receiving
// the messages can use that to rebuild the structure in a tree like manner/do anything that requires understanding the data in terms of a tree like structure
//
// right now, the only way to test the functionality in this file is to run an
// integration test that includes the front end where @run_tests.tsx the front end has two
// objects that are a reflection of how a query view that gets updated via an
// event tree and make sure that the two of them are in sync. However, it would
// be a nice to have a unit test for this too, as it will then be able to be ran
// more often and quicker.

package event_emitter_tree

import (
	"sql-compiler/compiler/rowType"
	pubsub "sql-compiler/pub_sub"
	"sql-compiler/utils"
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

type EventEmitterTree struct {
	On_message func(SyncMessage)
}

func (receiver *EventEmitterTree) SyncFromObservable(obs pubsub.ObservableI, path string) {
	if gb, ok := obs.(*pubsub.GroupBy); ok {
		receiver.SyncFromGroupByWithPathing(gb, path)
		return
	}
	obs.Add_sub(&pubsub.CustomSubscriber{
		OnAddFunc: func(item rowType.RowType) {
			primary_key := utils.String_or_num_to_string(item[0])
			receiver.On_message(SyncMessage{Type: SyncTypeAdd, Data: pubsub.RowTypeToJson(&item, obs.GetRowSchema()), Path: path + path_separator + primary_key})
			receiver.syncFromObservable_row(item, path+path_separator+primary_key, obs.GetRowSchema())
		},
		OnRemoveFunc: func(item rowType.RowType) {
			primary_key := utils.String_or_num_to_string(item[0])
			receiver.On_message(SyncMessage{Type: SyncTypeRemove, Data: pubsub.RowTypeToJson(&item, obs.GetRowSchema()), Path: path + path_separator + primary_key})
		},
		OnUpdateFunc: func(oldItem, newItem rowType.RowType) {
			primary_key := oldItem[0].(string)
			receiver.On_message(SyncMessage{Type: SyncTypeUpdate, Data: pubsub.RowTypeToJson(&newItem, obs.GetRowSchema()), Path: path + path_separator + primary_key})
		},
	})
	for row := range obs.Pull {
		receiver.syncFromObservable_row(row, path, obs.GetRowSchema())
	}

}

// this is more advanced, first start with @syncFromObservable_row and understand that, once you do and understand the idea of what were doing with the GroupBy class and what were trying to do to make a using a group by way more efficient then just doing subqueries then proceed to read this method
func (receiver *EventEmitterTree) SyncFromGroupByWithPathing(obs *pubsub.GroupBy, path string) {
	obs.Add_sub(&pubsub.CustomSubscriber{
		OnAddFunc: func(item rowType.RowType) {
			primary_key := utils.String_or_num_to_string(item[0])
			item_path := path + path_separator + obs.Get_rows_group_value(&item) + path_separator + primary_key
			receiver.On_message(SyncMessage{Type: SyncTypeAdd, Data: pubsub.RowTypeToJson(&item, obs.GetRowSchema()), Path: item_path})
			receiver.syncFromObservable_row(item, item_path, obs.GetRowSchema())
		},
		OnRemoveFunc: func(item rowType.RowType) {
			primary_key := utils.String_or_num_to_string(item[0])
			item_path := path + path_separator + obs.Get_rows_group_value(&item) + path_separator + primary_key
			receiver.On_message(SyncMessage{Type: SyncTypeRemove, Data: pubsub.RowTypeToJson(&item, obs.GetRowSchema()), Path: item_path})
		},
		OnUpdateFunc: func(oldItem, newItem rowType.RowType) {
			panic("todo: still working on this method")
			primary_key := utils.String_or_num_to_string(oldItem[0])
			item_path := path + path_separator + obs.Get_rows_group_value(&oldItem) + path_separator + primary_key
			receiver.On_message(SyncMessage{Type: SyncTypeUpdate, Data: pubsub.RowTypeToJson(&newItem, obs.GetRowSchema()), Path: item_path})
		},
	})
	for row := range obs.Pull {

		receiver.syncFromObservable_row(row, path+path_separator+obs.Get_rows_group_value(&row), obs.GetRowSchema())
	}

}

func (receiver *EventEmitterTree) syncFromObservable_row(row rowType.RowType, path string, row_schema rowType.RowSchema) {
	for i, col := range row {
		switch col := col.(type) {
		case string, int, bool:
		case pubsub.ObservableI:
			switch col := col.(type) {
			case *pubsub.Mapper:
				receiver.SyncFromObservable(col, path+path_separator+row_schema[i].Name)
			case *pubsub.GroupBy:
				receiver.SyncFromGroupByWithPathing(col, path+path_separator+row_schema[i].Name)
			default:
				panic("should be mapper")
			}
		default:
			panic("unhandled")
		}
	}
}
