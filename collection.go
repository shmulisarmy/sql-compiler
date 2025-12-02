package main

type R_Table struct {
	Observable
	rows       []RowType
	is_deleted []bool //use index to find out if the row at that index is deleted
}

func (this *R_Table) pull(yield func(RowType) bool) {
	for i, row := range this.rows {
		if !this.is_deleted[i] {
			yield(row)
		}
	}
}

func (this *R_Table) add(row RowType) {
	this.rows = append(this.rows, row)
	this.is_deleted = append(this.is_deleted, false)
	this.publish_add(row)
}

// //
// type Index struct {
// 	channels map[int]Channel
// }
