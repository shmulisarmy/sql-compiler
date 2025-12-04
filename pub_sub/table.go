package pubsub

import (
	. "sql-compiler/rowType"
	"sql-compiler/utils"
)

type R_Table struct {
	Observable
	rows       []RowType
	is_deleted []bool //use index to find out if the row at that index is deleted
	Indexes    []Index
}

func New_R_Table() R_Table {
	return R_Table{
		Observable: Observable{
			Subscribers: []Subscriber{},
		},
		rows:       []RowType{},
		is_deleted: []bool{},
		Indexes:    []Index{},
	}
}

func (this *R_Table) Pull(yield func(RowType) bool) {
	for i, row := range this.rows {
		if !this.is_deleted[i] {
			if !yield(row) {
				return
			}
		}
	}
}

func (this *R_Table) Add(row RowType) {
	this.rows = append(this.rows, row)
	this.is_deleted = append(this.is_deleted, false)
	///
	for i := range this.Indexes {
		channel_value := utils.String_or_num_to_string(row[this.Indexes[i].Col_indexing_on])
		if _, ok := this.Indexes[i].Channels[channel_value]; !ok {
			this.Indexes[i].Channels[channel_value] = NewChannel(this)
		}
		this.Indexes[i].Channels[channel_value].row_indexes = append(this.Indexes[i].Channels[channel_value].row_indexes, len(this.rows)-1)
		this.Indexes[i].Channels[channel_value].Publish_Add(row)
	}
	///
	this.Publish_Add(row)
}

// ///

type Index struct {
	Col_indexing_on int
	Channels        map[string]*Channel
	table           *R_Table
}

func (this *Index) Get_or_create_channel(row RowType) *Channel {
	if _, ok := this.Channels[row[this.Col_indexing_on].(string)]; !ok {
		this.Channels[row[this.Col_indexing_on].(string)] = NewChannel(this.table)
	}
	return this.Channels[row[this.Col_indexing_on].(string)]
}
func (this *Index) Get_or_create_channel_not_with_row(value string) *Channel {
	if _, ok := this.Channels[value]; !ok {
		this.Channels[value] = NewChannel(this.table)
	}
	return this.Channels[value]
}

func NewChannel(table *R_Table) *Channel {
	return &Channel{
		row_indexes: []int{},
		Observable: Observable{
			Subscribers: []Subscriber{},
		},
		table: table,
	}
}
func NewIndex(col_indexing_on int, table *R_Table) Index {
	return Index{
		Col_indexing_on: col_indexing_on,
		Channels:        map[string]*Channel{},
		table:           table,
	}
}

type Channel struct {
	row_indexes []int
	Observable
	table *R_Table //i want to remove the need to have this field by not using a generic pull, but rather use a pull method that takes in a reference to the table
}

func (this *Channel) Pull(yield func(RowType) bool) {
	for _, row_index := range this.row_indexes {
		if !yield(this.table.rows[row_index]) {
			return
		}
	}
}
