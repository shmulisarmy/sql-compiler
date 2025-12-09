package pubsub

import (
	"sql-compiler/compiler/rowType"
	"sql-compiler/utils"
)

type R_Table struct {
	Observable
	Rows       []rowType.RowType
	is_deleted []bool //use index to find out if the row at that index is deleted
	Indexes    []Index
	row_schema rowType.RowSchema
}

func New_R_Table(row_schema rowType.RowSchema) R_Table {
	return R_Table{
		Observable: Observable{
			Subscribers: []Subscriber{},
		},
		Rows:       []rowType.RowType{},
		is_deleted: []bool{},
		Indexes:    []Index{},
		row_schema: row_schema,
	}
}

func (this *R_Table) Pull(yield func(rowType.RowType) bool) {
	for i, row := range this.Rows {
		if !this.is_deleted[i] {
			if !yield(row) {
				return
			}
		}
	}
}

func (this *R_Table) Add(row rowType.RowType) {
	this.Rows = append(this.Rows, row)
	this.is_deleted = append(this.is_deleted, false)
	///
	for i := range this.Indexes {
		channel_value := utils.String_or_num_to_string(row[this.Indexes[i].Col_indexing_on])
		if _, ok := this.Indexes[i].Channels[channel_value]; !ok {
			this.Indexes[i].Channels[channel_value] = NewChannel(this)
		}
		this.Indexes[i].Channels[channel_value].row_indexes = append(this.Indexes[i].Channels[channel_value].row_indexes, len(this.Rows)-1)
		this.Indexes[i].Channels[channel_value].Publish_Add(row)
	}
	///
	this.Publish_Add(row)
}

// this is more for testing purposes because when integrating with the actual database (receiving and reacting to update events wel'e be updating by id)
func (this *R_Table) Remove_where_eq(row_schema rowType.RowSchema, field string, value any) {
	array_index := this.find_row_index(row_schema, field, value)
	if array_index == -1 {
		panic("not found")
	}
	this.is_deleted[array_index] = true
	this.Publish_remove(this.Rows[array_index])
}

// this is more for testing purposes because when integrating with the actual database (receiving and reacting to update events wel'e be updating by id)
func (this *R_Table) Update_where_eq(row_schema rowType.RowSchema, field string, value any, new_row rowType.RowType) {
	array_index := this.find_row_index(row_schema, field, value)
	if array_index == -1 {
		panic("not found")
	}
	old_row := this.Rows[array_index]
	this.Rows[array_index] = new_row
	this.Publish_Update(old_row, new_row)
}

func (this *R_Table) find_row_index(row_schema rowType.RowSchema, field string, value any) int {

	// look through the rows using the indexes
	for i := range this.Indexes {
		if this.Indexes[i].Col_indexing_on == row_schema.Find_field_index(field) {
			if channel, ok := this.Indexes[i].Channels[utils.String_or_num_to_string(value)]; ok {
				// assert.AssertEq(len(channel.row_indexes), 1)
				return channel.row_indexes[0]
			}
		}

	}

	// look through the rows manually
	for i := range this.Rows {
		if !this.is_deleted[i] {
			if this.Rows[i][row_schema.Find_field_index(field)] == value {
				return i
			}
		}
	}

	return -1
}

// ///

type Index struct {
	Col_indexing_on int
	Channels        map[string]*Channel
	table           *R_Table
}

func (this *Index) Get_or_create_channel(row rowType.RowType) *Channel {
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

func (this *Channel) Pull(yield func(rowType.RowType) bool) {
	for _, row_index := range this.row_indexes {
		if !yield(this.table.Rows[row_index]) {
			return
		}
	}
}

func (this *R_Table) GetRowSchema() rowType.RowSchema {
	return this.row_schema
}

func (this *Channel) GetRowSchema() rowType.RowSchema {
	return this.table.row_schema
}
