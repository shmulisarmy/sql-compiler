package main

func (this *R_Table) filter_on(predicate func(RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	link(this, f)
	return f
}

func (this *R_Table) map_on(transformer func(RowType) RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	link(this, m)
	return m
}

func (this *R_Table) to_display() *Printer {
	p := &Printer{}
	link(this, p)
	return p
}

// ///
func (this *Mapper) filter_on(predicate func(RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	link(this, f)
	return f
}

func (this *Mapper) map_on(transformer func(RowType) RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	link(this, m)
	return m
}

func (this *Mapper) to_display() *Printer {
	p := &Printer{}
	link(this, p)
	return p
}

/////

func (this *Filter) filter_on(predicate func(RowType) bool) ObservableI {
	f := &Filter{predicate: predicate}
	link(this, f)
	return f
}

func (this *Filter) map_on(transformer func(RowType) RowType) ObservableI {
	m := &Mapper{transformer: transformer}
	link(this, m)
	return m
}

func (this *Filter) to_display() *Printer {
	p := &Printer{}
	link(this, p)
	return p
}

////
