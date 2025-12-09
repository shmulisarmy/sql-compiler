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
