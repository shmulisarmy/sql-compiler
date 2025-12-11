package pubsub

import (
	"sql-compiler/compiler/rowType"
	"sql-compiler/unwrap"
)

func (this *R_Table) Filter_on(predicate func(rowType.RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	Link(this, f)
	return f
}

func (this *R_Table) Map_on(transformer func(rowType.RowType) rowType.RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	Link(this, m)
	return m
}

func (this *R_Table) To_display(row_schema unwrap.Option[rowType.RowSchema]) *Printer {
	p := &Printer{
		RowSchema: row_schema,
	}
	Link(this, p)
	return p
}

func (this *R_Table) GroupBy_on(col_index int) ObservableI {
	g := &GroupBy{index_of_col_to_group_by: col_index}
	Link(this, g)
	return g
}

// ///
func (this *Mapper) Filter_on(predicate func(rowType.RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	Link(this, f)
	return f
}

func (this *Mapper) Map_on(transformer func(rowType.RowType) rowType.RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	Link(this, m)
	return m
}

func (this *Mapper) To_display(row_schema unwrap.Option[rowType.RowSchema]) *Printer {
	p := &Printer{
		RowSchema: row_schema,
	}
	Link(this, p)
	return p
}

func (this *Mapper) GroupBy_on(col_index int) ObservableI {
	g := &GroupBy{index_of_col_to_group_by: col_index}
	Link(this, g)
	return g
}

/////

func (this *Filter) Filter_on(predicate func(rowType.RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	Link(this, f)
	return f
}

func (this *Filter) Map_on(transformer func(rowType.RowType) rowType.RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	Link(this, m)
	return m
}

func (this *Filter) To_display(row_schema unwrap.Option[rowType.RowSchema]) *Printer {
	p := &Printer{
		RowSchema: row_schema,
	}
	Link(this, p)
	return p
}

func (this *Filter) GroupBy_on(col_index int) ObservableI {
	g := &GroupBy{index_of_col_to_group_by: col_index}
	Link(this, g)
	return g
}

////

func (this *Channel) Filter_on(predicate func(rowType.RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	Link(this, f)
	return f
}

func (this *Channel) Map_on(transformer func(rowType.RowType) rowType.RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	Link(this, m)
	return m
}

func (this *Channel) To_display(row_schema unwrap.Option[rowType.RowSchema]) *Printer {
	p := &Printer{
		RowSchema: row_schema,
	}
	Link(this, p)
	return p
}

func (this *Channel) GroupBy_on(col_index int) ObservableI {
	g := &GroupBy{index_of_col_to_group_by: col_index}
	Link(this, g)
	return g
}

///

func (this *CustomSubscriber) Filter_on(predicate func(rowType.RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	Link(this, f)
	return f
}

func (this *CustomSubscriber) Map_on(transformer func(rowType.RowType) rowType.RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	Link(this, m)
	return m
}

func (this *CustomSubscriber) To_display(row_schema unwrap.Option[rowType.RowSchema]) *Printer {
	p := &Printer{
		RowSchema: row_schema,
	}
	Link(this, p)
	return p
}

func (this *CustomSubscriber) GroupBy_on(col_index int) ObservableI {
	g := &GroupBy{index_of_col_to_group_by: col_index}
	Link(this, g)
	return g
}

///

func (this *FullOuterJoin) Filter_on(predicate func(rowType.RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	Link(this, f)
	return f
}

func (this *FullOuterJoin) Map_on(transformer func(rowType.RowType) rowType.RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	Link(this, m)
	return m
}

func (this *FullOuterJoin) To_display(row_schema unwrap.Option[rowType.RowSchema]) *Printer {
	p := &Printer{
		RowSchema: row_schema,
	}
	Link(this, p)
	return p
}

func (this *FullOuterJoin) GroupBy_on(col_index int) ObservableI {
	g := &GroupBy{index_of_col_to_group_by: col_index}
	Link(this, g)
	return g
}

//

func (this *GroupBy) Filter_on(predicate func(rowType.RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	Link(this, f)
	return f
}

func (this *GroupBy) Map_on(transformer func(rowType.RowType) rowType.RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	Link(this, m)
	return m
}

func (this *GroupBy) To_display(row_schema unwrap.Option[rowType.RowSchema]) *Printer {
	p := &Printer{
		RowSchema: row_schema,
	}
	Link(this, p)
	return p
}

func (this *GroupBy) GroupBy_on(col_index int) ObservableI {
	g := &GroupBy{index_of_col_to_group_by: col_index}
	Link(this, g)
	return g
}
