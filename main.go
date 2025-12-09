package main

import (
	"net/http"
	"os"
	compiler_runtime "sql-compiler/compiler/runtime"
	"sql-compiler/db_tables"
	"sql-compiler/display"
	pubsub "sql-compiler/pub_sub"
	"sql-compiler/rowType"
	. "sql-compiler/rowType"

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

	obs := compiler_runtime.Query_to_observer(src)

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

func add_sample_data() {
	tables := db_tables.Tables
	tables.Get("person").Insert(rowType.RowType{"shmuli", "email@gmail.com", 25, "state", tables.Get("person").Next_row_id()})
	tables.Get("person").Insert(rowType.RowType{"the-doo-er", "email@gmail.com", 20, "state", tables.Get("person").Next_row_id()})
	todos_table.Insert(rowType.RowType{"eat food", "make sure its clean", false, 1, false})
	todos_table.Insert(rowType.RowType{"play music", "make sure its clean", false, 1, true})
	todos_table.Insert(rowType.RowType{"clean", "make sure its clean", true, 1, false})
	todos_table.Insert(rowType.RowType{"do art", "make sure its clean", false, 2, true})
}
