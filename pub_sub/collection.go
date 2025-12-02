package pubsub

import (
	. "sql-compiler/rowType"
)

type R_Table struct {
	Observable
	rows       []RowType
	is_deleted []bool //use index to find out if the row at that index is deleted
}

func New_R_Table() R_Table {
	return R_Table{
		Observable: Observable{
			Subscribers: []Subscriber{},
		},
		rows:       []RowType{},
		is_deleted: []bool{},
	}
}

func (this *R_Table) Pull(yield func(RowType) bool) {
	for i, row := range this.rows {
		if !this.is_deleted[i] {
			yield(row)
		}
	}
}

func (this *R_Table) Add(row RowType) {
	this.rows = append(this.rows, row)
	this.is_deleted = append(this.is_deleted, false)
	this.Publish_Add(row)
}

// //
// type Index struct {
// 	channels map[int]Channel
// }
