package db_tables

import (
	"fmt"
	"sql-compiler/assert"
	"sql-compiler/compiler/rowType"
	"sql-compiler/display"
	pubsub "sql-compiler/pub_sub"
	"sql-compiler/utils"
)

type Table struct {
	Name    string
	Columns []rowType.ColInfo
	R_Table pubsub.R_Table
}

func NewTable(name string, columns []rowType.ColInfo) Table {
	return Table{
		Name:    name,
		Columns: columns,
		R_Table: pubsub.New_R_Table(columns),
	}
}

func (this *Table) Next_row_id() int {
	return len(this.R_Table.Rows)
}

func (this *Table) HasCol(col_name string) bool {
	for i := range this.Columns {
		if this.Columns[i].Name == col_name {
			return true
		}
	}
	return false
}

func (this *Table) HasIndex(col_name string) bool {
	for i := range this.R_Table.Indexes {
		if this.R_Table.Indexes[i].Col_indexing_on == this.Get_col_index(col_name) {
			return true
		}
	}
	return false
}

func (this *Table) Index_on(col_name string) *pubsub.Index {
	for i := range this.R_Table.Indexes {
		if this.R_Table.Indexes[i].Col_indexing_on == this.Get_col_index(col_name) {
			return &this.R_Table.Indexes[i]
		}
	}
	display.DisplayStruct(this)
	this.R_Table.Indexes = append(this.R_Table.Indexes, pubsub.NewIndex(this.Get_col_index(col_name), &this.R_Table))
	display.DisplayStruct(this)
	return &this.R_Table.Indexes[len(this.R_Table.Indexes)-1]
}

func (this *Table) Insert(row rowType.RowType) {
	assert.AssertEq(len(row), len(this.Columns), fmt.Sprintf("rows in table %s must have %d columns and you passed a row that has %d columns", this.Name, len(this.Columns), len(row)))
	validate_col_types(this, &row)
	this.R_Table.Add(row)
}

func validate_col_types(this *Table, row *rowType.RowType) {
	for i, col := range this.Columns {
		switch col.Type {
		case rowType.String:
			if _, ok := (*row)[i].(string); !ok {
				panic(fmt.Sprintf("col %s of table %s's type is string and you passed in a %T", col.Name, this.Name, (*row)[i]))
			}
		case rowType.Int:
			if _, ok := (*row)[i].(int); !ok {
				panic(fmt.Sprintf("col %s of table %s's type is int and you passed in a %T", col.Name, this.Name, (*row)[i]))
			}
		case rowType.Bool:
			if _, ok := (*row)[i].(bool); !ok {
				panic(fmt.Sprintf("col %s of table %s's type is bool and you passed in a %T", col.Name, this.Name, (*row)[i]))
			}
		default:
			panic("unhandled")
		}
	}
}

func (this Table) Get_index(col_name string) *pubsub.Index {
	for i := range this.R_Table.Indexes {
		if this.R_Table.Indexes[i].Col_indexing_on == this.Get_col_index(col_name) {
			return &this.R_Table.Indexes[i]
		}
	}
	panic("col " + col_name + " not found in table " + this.Name)
}

func (this Table) Get_col_index(col_name string) int {
	for i, col := range this.Columns {
		if col.Name == col_name {
			return i
		}
	}
	return -1
}

var Tables = tablesNewKeyValueArrayWith(30, NewTable("person", rowType.RowSchema{{"name", rowType.String}, {"email", rowType.String}, {"age", rowType.Int}, {"state", rowType.String}, {"id", rowType.Int}}),
	NewTable("todo", []rowType.ColInfo{{"title", rowType.String}, {"description", rowType.String}, {"done", rowType.Bool}, {"person_id", rowType.Int}, {"is_public", rowType.Bool}}))

func tablesNewKeyValueArrayWith(constant_cap int, initial_tables ...Table) *utils.CappedKeyValueArray[Table] {
	keyValueArray := utils.NewKeyValueArray[Table](constant_cap)
	for _, table := range initial_tables {
		keyValueArray.Add(table.Name, table)
	}
	return keyValueArray
}
