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
