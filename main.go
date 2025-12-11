package main

import (
	"net/http"
	"sql-compiler/compiler/rowType"
	compiler_runtime "sql-compiler/compiler/runtime"
	"sql-compiler/db_tables"
	"sql-compiler/display"
	event_emitter_tree "sql-compiler/eventEmitterTree"
	pubsub "sql-compiler/pub_sub"

	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var todos_table *db_tables.Table

func init() {
	todos_table = db_tables.Tables.Get("todo")
}

func obsToClientDataSync(obs pubsub.ObservableI, ws *websocket.Conn) {
	eventEmitterTree := event_emitter_tree.EventEmitterTree{
		On_message: func(message event_emitter_tree.SyncMessage) {
			message.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
			ws.WriteJSON(message)
		},
	}
	eventEmitterTree.SyncFromObservable(obs, "")
	eventEmitterTree.On_message(event_emitter_tree.SyncMessage{Type: event_emitter_tree.LoadInitialData, Data: pubsub.ObserverToJson(obs, obs.GetRowSchema())})
}

func add_sample_data() {
	tables := db_tables.Tables
	tables.Get("person").Insert(rowType.RowType{"shmuli", "email@gmail.com", 22, "state", tables.Get("person").Next_row_id()})
	tables.Get("person").Insert(rowType.RowType{"ajay", "email@gmail.com", 22, "state", tables.Get("person").Next_row_id()})
	tables.Get("person").Insert(rowType.RowType{"the-doo-er", "email@gmail.com", 20, "state", tables.Get("person").Next_row_id()})
	todos_table.Insert(rowType.RowType{"eat food", "make sure its clean", false, 1, false})
	todos_table.Insert(rowType.RowType{"play music", "make sure its clean", false, 1, true})
	todos_table.Insert(rowType.RowType{"clean", "make sure its clean", true, 1, false})
	todos_table.Insert(rowType.RowType{"do art", "make sure its clean", false, 2, true})
}

func main() {

	// gin.SetMode("release")
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

	src := `SELECT person.name, person.email, person.age, person.id, (
		SELECT todo.title as epic_title, person.name as author, person.id FROM todo WHERE todo.person_id == person.id
		) as todos FROM person WHERE person.age >= 3  `

	obs := compiler_runtime.Query_to_observer(src)

	display.DisplayStruct(obs)

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
	eventEmitterTree := event_emitter_tree.EventEmitterTree{
		On_message: func(message event_emitter_tree.SyncMessage) {
			display.DisplayStruct(message)
		},
	}
	eventEmitterTree.SyncFromObservable(obs, "")
	r.GET("add-sample-data", func(ctx *gin.Context) {
		add_sample_data()
	})
	r.Run(":8080")

	// os.Exit(0)

}
