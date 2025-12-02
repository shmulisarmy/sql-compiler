package pubsub

import (
	"fmt"
	. "sql-compiler/rowType"
)

type Printer struct {
	Observable
	subscribed_to ObservableI
}

func (this *Printer) set_subscribed_to(observable ObservableI) {
	this.subscribed_to = observable
}

func (this *Printer) on_Add(row RowType) {
	fmt.Println("Added row ", row)
}

func (this *Printer) on_remove(row RowType) {
	fmt.Println("removed row ", row)
}

func (this *Printer) on_update(old_row RowType, new_row RowType) {
	fmt.Println("updated row from ", old_row, "to", new_row)
}

func (this *Printer) run() {
	this.subscribed_to.Pull(func(row RowType) bool {
		fmt.Println("row ", row)
		return true
	})
}
