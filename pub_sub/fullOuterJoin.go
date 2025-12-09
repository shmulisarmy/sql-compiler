package pubsub

import (
	"fmt"
	"sql-compiler/compiler/rowType"
	"sql-compiler/utils"
)

type Tuple[T any] struct {
	first  T
	second T
}

type TypedObservableI interface {
	ObservableI
	GetRowSchema() rowType.RowSchema
}

type FullOuterJoin struct {
	Observable
	source_one_on     int
	source_two_on     int
	source_one        TypedObservableI
	source_two        TypedObservableI
	values            map[string]Tuple[*[]rowType.RowType]
	output_row_schema rowType.RowSchema
}

func combineSchemas(schema1, schema2 rowType.RowSchema) rowType.RowSchema {
	result := make(rowType.RowSchema, len(schema1)+len(schema2))
	copy(result, schema1)
	copy(result[len(schema1):], schema2)
	return result
}

func NewFullOuterJoin(source_one TypedObservableI, source_two TypedObservableI, source_one_on int, source_two_on int) *FullOuterJoin {

	res := combineSchemas(source_one.GetRowSchema(), source_two.GetRowSchema())
	for _, col_info := range res {
		fmt.Printf("%s: %s\n", col_info.Name, col_info.Type.To_string(0))
	}
	j := &FullOuterJoin{
		Observable: Observable{
			Subscribers: []Subscriber{},
		},
		source_one:        source_one,
		source_two:        source_two,
		source_one_on:     source_one_on,
		source_two_on:     source_two_on,
		output_row_schema: res,
		values:            make(map[string]Tuple[*[]rowType.RowType]),
	}

	Link(source_one, &CustomSubscriber{
		OnAddFunc:    j.source_one_on_Add,
		OnRemoveFunc: j.source_one_on_Remove,
		OnUpdateFunc: j.source_one_on_update,
	})
	Link(source_two, &CustomSubscriber{
		OnAddFunc:    j.source_two_on_Add,
		OnRemoveFunc: j.source_two_on_Remove,
		OnUpdateFunc: j.source_two_on_update,
	})
	return j
}

func (this *FullOuterJoin) combine_rows(row1 rowType.RowType, row2 rowType.RowType) rowType.RowType {
	result := make(rowType.RowType, len(row1)+len(row2))
	copy(result, row1)
	copy(result[len(row1):], row2)
	return result
}
func (this *FullOuterJoin) Pull(yield func(rowType.RowType) bool) {
	for col_value, rows := range this.values {
		for _, row1 := range *rows.first {
			if _, ok := this.values[col_value]; !ok {
				this.values[col_value] = Tuple[*[]rowType.RowType]{
					first:  &[]rowType.RowType{},
					second: &[]rowType.RowType{},
				}
			}
			for _, row2 := range *this.values[col_value].second {
				if !yield(this.combine_rows(row1, row2)) {
					return
				}

			}
		}
	}
}

func (this *FullOuterJoin) source_one_on_Add(row1 rowType.RowType) {
	// this.Publish_Add(row1)
	col_value := row1[this.source_one_on]
	if _, ok := this.values[col_value.(string)]; !ok {
		this.values[col_value.(string)] = Tuple[*[]rowType.RowType]{
			first:  &[]rowType.RowType{},
			second: &[]rowType.RowType{},
		}
	}
	for _, row2 := range *this.values[col_value.(string)].second {
		this.Publish_Add(this.combine_rows(row1, row2))

	}
	(*this.values[col_value.(string)].first) = append(*this.values[col_value.(string)].first, row1)
}

func (this *FullOuterJoin) source_two_on_Add(row2 rowType.RowType) {
	// this.Publish_Add(row)
	col_value := row2[this.source_two_on]
	if _, ok := this.values[col_value.(string)]; !ok {
		this.values[col_value.(string)] = Tuple[*[]rowType.RowType]{
			first:  &[]rowType.RowType{},
			second: &[]rowType.RowType{},
		}
	}
	for _, row1 := range *this.values[col_value.(string)].first {
		this.Publish_Add(this.combine_rows(row1, row2))

	}
	(*this.values[col_value.(string)].second) = append(*this.values[col_value.(string)].second, row2)
}

func (this *FullOuterJoin) source_one_on_Remove(row rowType.RowType) {
	col_value := row[this.source_one_on]
	if _, ok := this.values[col_value.(string)]; !ok {
		this.values[col_value.(string)] = Tuple[*[]rowType.RowType]{
			first:  &[]rowType.RowType{},
			second: &[]rowType.RowType{},
		}
	}
	for i, row1 := range *this.values[col_value.(string)].first {
		if !utils.CompareSlices(row1, row) {
			continue
		}
		for _, row2 := range *this.values[col_value.(string)].second {
			this.Publish_remove(this.combine_rows(row1, row2))
		}
		(*this.values[col_value.(string)].first) = append((*this.values[col_value.(string)].first)[:i], (*this.values[col_value.(string)].first)[i+1:]...)

	}
}

func (this *FullOuterJoin) source_two_on_Remove(row rowType.RowType) {
	col_value := row[this.source_two_on]
	if _, ok := this.values[col_value.(string)]; !ok {
		this.values[col_value.(string)] = Tuple[*[]rowType.RowType]{
			first:  &[]rowType.RowType{},
			second: &[]rowType.RowType{},
		}
	}
	for i, row1 := range *this.values[col_value.(string)].second {
		if !utils.CompareSlices(row1, row) {
			continue
		}
		for _, row2 := range *this.values[col_value.(string)].first {
			this.Publish_remove(this.combine_rows(row1, row2))
		}
		(*this.values[col_value.(string)].second) = append((*this.values[col_value.(string)].second)[:i], (*this.values[col_value.(string)].second)[i+1:]...)

	}
}

func (this *FullOuterJoin) source_one_on_update(old_row rowType.RowType, new_row rowType.RowType) {
	col_value := old_row[this.source_one_on]
	if _, ok := this.values[col_value.(string)]; !ok {
		this.values[col_value.(string)] = Tuple[*[]rowType.RowType]{
			first:  &[]rowType.RowType{},
			second: &[]rowType.RowType{},
		}
	}
	for i, row := range *this.values[col_value.(string)].first {
		if !utils.CompareSlices(old_row, row) {
			continue
		}
		for _, row1 := range *this.values[col_value.(string)].first {
			this.Publish_Update(this.combine_rows(old_row, row), this.combine_rows(new_row, row1))
		}
		(*this.values[col_value.(string)].first) = append((*this.values[col_value.(string)].first)[:i], (*this.values[col_value.(string)].first)[i+1:]...)

	}
}

func (this *FullOuterJoin) source_two_on_update(old_row rowType.RowType, new_row rowType.RowType) {
	col_value := old_row[this.source_one_on]
	if _, ok := this.values[col_value.(string)]; !ok {
		this.values[col_value.(string)] = Tuple[*[]rowType.RowType]{
			first:  &[]rowType.RowType{},
			second: &[]rowType.RowType{},
		}
	}
	for i, row := range *this.values[col_value.(string)].first {
		if !utils.CompareSlices(row, old_row) {
			continue
		}
		for _, row1 := range *this.values[col_value.(string)].first {
			this.Publish_Update(this.combine_rows(row, old_row), this.combine_rows(row1, new_row))
		}
		(*this.values[col_value.(string)].second) = append((*this.values[col_value.(string)].second)[:i], (*this.values[col_value.(string)].second)[i+1:]...)

	}
}
func (this *FullOuterJoin) GetRowSchema() rowType.RowSchema {
	return this.output_row_schema

}
