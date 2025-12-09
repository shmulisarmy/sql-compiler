package main

import (
	"fmt"
	"net/http"
	"os"
	"sql-compiler/byte_code"
	"sql-compiler/db_tables"
	"sql-compiler/display"
	. "sql-compiler/parser"
	. "sql-compiler/parser/tokenizer"
	pubsub "sql-compiler/pub_sub"
	"sql-compiler/rowType"
	. "sql-compiler/rowType"
	"sql-compiler/state_full_byte_code"
	option "sql-compiler/unwrap"
	"sql-compiler/utils"
	. "sql-compiler/utils"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func init() {
	db_tables.Tables.Add("person", db_tables.Table{
		Name:    "person",
		Columns: []ColInfo{{"name", String}, {"email", String}, {"age", Int}, {"state", String}, {"id", Int}},
		R_Table: pubsub.New_R_Table(),
	})
	db_tables.Tables.Add("todo", db_tables.Table{
		Name:    "todo",
		Columns: []ColInfo{{"title", String}, {"description", String}, {"done", Bool}, {"person_id", Int}, {"is_public", Bool}},
		R_Table: pubsub.New_R_Table(),
	})
}

var compare_methods = map[string]func(value1 any, value2 any) bool{

	"==": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 == value2.(string)
		case int:
			return value1 == value2.(int)
		case bool:
			return value1 == value2.(bool)
		default:
			panic(fmt.Sprintf("types %T and %T do not match", value1, value2))
		}
	},
	">": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 > value2.(string)
		case int:
			return value1 > value2.(int)
		case bool:
			return value1 == value2.(bool)
		default:
			panic(fmt.Sprintf("types %T and %T do not match", value1, value2))
		}
	},
	"<": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 < value2.(string)
		case int:
			return value1 < value2.(int)
		case bool:
			return value1 == value2.(bool)
		default:
			panic(fmt.Sprintf("types %T and %T do not match", value1, value2))
		}
	},
	">=": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 >= value2.(string)
		case int:
			return value1 >= value2.(int)
		case bool:
			return value1 == value2.(bool)
		default:
			panic(fmt.Sprintf("types %T and %T do not match", value1, value2))
		}
	},
	"<=": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 <= value2.(string)
		case int:
			return value1 <= value2.(int)
		case bool:
			return value1 == value2.(bool)
		default:
			panic(fmt.Sprintf("types %T and %T do not match", value1, value2))
		}
	},
}

func filter(row_context state_full_byte_code.Row_context, wheres []byte_code.Where) bool {
	for _, where := range wheres {
		if !compare_methods[where.Compare_type](row_context.Track_value_if_is_relative_location(where.Value_1), row_context.Track_value_if_is_relative_location(where.Value_2)) {
			return false
		}
	}
	return true
}

func map_over(row_context state_full_byte_code.Row_context, selected_values_byte_code []byte_code.Expression, row_schema rowType.RowSchema) rowType.RowType {
	row := rowType.RowType{}
	for i, select_value_byte_code := range selected_values_byte_code { ///select_value_byte_code could just be a plain value
		switch select_value_byte_code := select_value_byte_code.(type) {
		case byte_code.Runtime_value_relative_location:
			row = append(row, row_context.Get_value(select_value_byte_code))
		case byte_code.Select:
			childs_row_context := state_full_byte_code.Row_context{Row: row_context.Row, Parent_context: option.Some(&row_context)}
			childs_row_schema := NestedSelectsRowSchema[row_schema[i].Type]
			row = append(row, select_byte_code_to_observable(select_value_byte_code, option.Some(&childs_row_context), childs_row_schema))
		default:
			row = append(row, select_value_byte_code)
		}
	}
	return row
}

func select_byte_code_to_observable(select_byte_code byte_code.Select, parent_context option.Option[*state_full_byte_code.Row_context], row_schema rowType.RowSchema) *pubsub.Mapper {
	var current_observable pubsub.ObservableI
	if select_byte_code.Col_and_value_to_index_by.Col != "" {
		//ints are cast to strings when placed and queried from indexes
		channel_value := select_byte_code.Col_and_value_to_index_by.Value
		switch channel_value := channel_value.(type) {
		case byte_code.Runtime_value_relative_location:
			tracked_channel_value := parent_context.Unwrap().Get_value(channel_value)
			current_observable = db_tables.Tables.Get(select_byte_code.Table_name).Index_on(select_byte_code.Col_and_value_to_index_by.Col).Get_or_create_channel_not_with_row(String_or_num_to_string(tracked_channel_value))
		case string:
			current_observable = db_tables.Tables.Get(select_byte_code.Table_name).Index_on(select_byte_code.Col_and_value_to_index_by.Col).Get_or_create_channel_not_with_row(channel_value)
		case int:
			int_str := strconv.Itoa(channel_value)
			current_observable = db_tables.Tables.Get(select_byte_code.Table_name).Index_on(select_byte_code.Col_and_value_to_index_by.Col).Get_or_create_channel_not_with_row(int_str)
		default:
			//bools are not supported for indexing indexes
			panic(fmt.Sprintf("%T %s", channel_value, channel_value))
		}
	} else {
		current_observable = &db_tables.Tables.Get(select_byte_code.Table_name).R_Table
	}

	current_observable = current_observable.Filter_on(func(row rowType.RowType) bool {
		return filter(state_full_byte_code.Row_context{Row: row, Parent_context: parent_context}, select_byte_code.Wheres_byte_code)
	}).Map_on(func(row rowType.RowType) rowType.RowType {
		return map_over(state_full_byte_code.Row_context{Row: row, Parent_context: parent_context}, select_byte_code.Selected_values_byte_code, row_schema)
	})

	current_observable.(*pubsub.Mapper).RowSchema = option.Some(row_schema)
	return current_observable.(*pubsub.Mapper)

}

var todos_table *db_tables.Table

func init() {
	todos_table = db_tables.Tables.Get("todo")
}

func obsToClientDataSync(obs *pubsub.Mapper, ws *websocket.Conn) {
	eventEmitterTree := eventEmitterTree{
		on_message: func(message SyncMessage) {
			message.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
			ws.WriteJSON(message)
		},
	}
	eventEmitterTree.syncFromObservable(obs, "")
	eventEmitterTree.on_message(SyncMessage{Type: LoadInitialData, Data: pubsub.ObserverToJson(obs, obs.RowSchema.Expect("if this was a mapper that was made by compiling a select statement then it should have a row schema"))})
}
func main() {

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		AllowWebSockets:  true,
	}))

	db_tables.Tables.Get("person").Index_on("age")

	db_tables.Tables.Get("todo").Index_on("person_id")

	src := `SELECT person.name, person.email, person.id, (
		SELECT todo.title as epic_title, person.name as author, person.id FROM todo WHERE todo.person_id == person.id
		) as todo FROM person WHERE person.age >= 3 `

	obs := query_to_observer(src)

	r.GET("/stream-data", func(ctx *gin.Context) {
		ws, err := (&websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}).Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			panic(err)
		}
		obsToClientDataSync(obs, ws)
	})

	r.GET("add-person", func(ctx *gin.Context) {
		db_tables.Tables.Get("person").Insert(rowType.RowType{ctx.Query("name"), ctx.Query("email"), 25, "state", db_tables.Tables.Get("person").Next_row_id()})
	})
	r.GET("add-todo", func(ctx *gin.Context) {
		person_id, err := strconv.Atoi(ctx.Query("person_id"))
		if err != nil {
			panic(err)
		}

		db_tables.Tables.Get("todo").Insert(rowType.RowType{ctx.Query("title"), ctx.Query("description"), false, person_id, true})
	})
	eventEmitterTree := eventEmitterTree{
		on_message: func(message SyncMessage) {
			display.DisplayStruct(message)
		},
	}
	eventEmitterTree.syncFromObservable(obs, "")
	r.GET("add-sample-data", func(ctx *gin.Context) {
		add_sample_data()
	})
	r.Run(":8080")

	os.Exit(0)

}

func query_to_observer(src string) *pubsub.Mapper {
	l := NewLexer(src)
	parser := Parser{Tokens: l.Tokenize()}
	for _, t := range parser.Tokens {
		fmt.Printf("%-8s %q @%d\n", t.Type, t.Literal, t.Pos)
	}
	select_ := parser.Parse_Select()
	select_.Recursively_link_children()
	Recursively_set_selects_row_schema(&select_)
	display.DisplayStruct(select_)
	select_byte_code := make_select_byte_code(&select_)
	display.DisplayStruct(select_byte_code)

	obs := select_byte_code_to_observable(select_byte_code, option.None[*state_full_byte_code.Row_context](), select_.Row_schema)
	obs.To_display(option.Some(select_.Row_schema))
	fmt.Printf("type %s=%s\n", utils.Capitalize(select_.Table), select_.Row_schema.To_string(0))

	return obs
}

func add_sample_data() {
	tables := db_tables.Tables
	tables.Get("person").Insert(rowType.RowType{"shmuli", "email@gmail.com", 25, "state", tables.Get("person").Next_row_id()})
	tables.Get("person").Insert(rowType.RowType{"the-doo-er", "email@gmail.com", 20, "state", tables.Get("person").Next_row_id()})
	todos_table.Insert(rowType.RowType{"eat food", "make sure its clean", false, 1, false})
	todos_table.Insert(rowType.RowType{"play music", "make sure its clean", false, 1, true})
	todos_table.Insert(rowType.RowType{"clean", "make sure its clean", true, 1, false})
	todos_table.Insert(rowType.RowType{"do art", "make sure its clean", false, 2, true})
}
