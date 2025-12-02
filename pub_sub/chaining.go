package pubsub

import (
	. "sql-compiler/rowType"
)

func (this *R_Table) Filter_on(predicate func(RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	link(this, f)
	return f
}

func (this *R_Table) Map_on(transformer func(RowType) RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	link(this, m)
	return m
}

func (this *R_Table) To_display() *Printer {
	p := &Printer{}
	link(this, p)
	return p
}

// ///
func (this *Mapper) Filter_on(predicate func(RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	link(this, f)
	return f
}

func (this *Mapper) Map_on(transformer func(RowType) RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	link(this, m)
	return m
}

func (this *Mapper) To_display() *Printer {
	p := &Printer{}
	link(this, p)
	return p
}

/////

func (this *Filter) Filter_on(predicate func(RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	link(this, f)
	return f
}

func (this *Filter) Map_on(transformer func(RowType) RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	link(this, m)
	return m
}

func (this *Filter) To_display() *Printer {
	p := &Printer{}
	link(this, p)
	return p
}

////
