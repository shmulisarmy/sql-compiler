package pubsub

import (
	"fmt"
	"sql-compiler/rowType"
	"sql-compiler/unwrap"
)

type Printer struct {
	Observable
	subscribed_to ObservableI
	RowSchema     unwrap.Option[rowType.RowSchema]
}

func (this *Printer) set_subscribed_to(observable ObservableI) {
	this.subscribed_to = observable
}

func (this *Printer) on_Add(row rowType.RowType) {
	if this.RowSchema.IsSome() {
		fmt.Printf("Added row %s\n", RowTypeToJson(&row, this.RowSchema.Unwrap()))
	} else {
		fmt.Println("Added row ", row)
	}
}

func (this *Printer) on_remove(row rowType.RowType) {
	if this.RowSchema.IsSome() {
		fmt.Printf("removed row %s\n", RowTypeToJson(&row, this.RowSchema.Unwrap()))
	} else {
		fmt.Println("removed row ", row)
	}
}

func (this *Printer) on_update(old_row rowType.RowType, new_row rowType.RowType) {
	if this.RowSchema.IsSome() {
		fmt.Printf("updated row from %s to %s\n", RowTypeToJson(&old_row, this.RowSchema.Unwrap()), RowTypeToJson(&new_row, this.RowSchema.Unwrap()))
	} else {
		fmt.Printf("updated row from %v to %v\n", old_row, new_row)
	}
}

func (this *Printer) run() {
	for row := range this.subscribed_to.Pull {
		if this.RowSchema.IsSome() {
			fmt.Printf("row %s\n", RowTypeToJson(&row, this.RowSchema.Unwrap()))
		} else {
			fmt.Println("row ", row)
		}
	}
}
