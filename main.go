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

//go:embed all:frontend/dist
// var frontendFS embed.FS

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
	tables.Get("person").Insert(rowType.RowType{"teddy", "teddyemail@gmail.com", 22, "state", tables.Get("person").Next_row_id(), "https://api.dicebear.com/7.x/avataaars/svg?seed=teddy"})
	tables.Get("person").Insert(rowType.RowType{"ariana", "arianaemail@gmail.com", 22, "state", tables.Get("person").Next_row_id(), "https://api.dicebear.com/7.x/avataaars/svg?seed=ajay"})
	tables.Get("person").Insert(rowType.RowType{"the-doo-er", "the-doo-eremail@gmail.com", 20, "state", tables.Get("person").Next_row_id(), "https://api.dicebear.com/7.x/avataaars/svg?seed=the-doo-er"})
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

	src := `SELECT person.name, person.email, person.age, person.id, person.profile_picture FROM person WHERE person.age >= 3 `

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
		name := ctx.Query("name")
		profile_picture := ctx.Query("profile_picture")
		if profile_picture == "" {
			profile_picture = "https://api.dicebear.com/7.x/avataaars/svg?seed=" + name
		}
		db_tables.Tables.Get("person").Insert(rowType.RowType{name, ctx.Query("email"), 25, "state", db_tables.Tables.Get("person").Next_row_id(), profile_picture})
	})
	r.GET("delete-person", func(ctx *gin.Context) {
		person_id, err := strconv.Atoi(ctx.Query("id"))
		if err != nil {
			panic(err)
		}
		person_table := db_tables.Tables.Get("person")
		row_schema := rowType.RowSchema(person_table.Columns)
		person_table.R_Table.Remove_where_eq(row_schema, "id", person_id)
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

	// r.Use(func(c *gin.Context) {
	// 	path := c.Request.URL.Path

	// 	if strings.HasPrefix(path, "/stream-data") || strings.HasPrefix(path, "/add-person") || strings.HasPrefix(path, "/delete-person") || strings.HasPrefix(path, "/add-sample-data") {
	// 		c.Next()
	// 		return
	// 	}

	// 	filePath := strings.TrimPrefix(path, "/")
	// 	_, err := frontendDist.Open(filePath)

	// 	if err != nil {
	// 		c.FileFromFS("index.html", http.FS(frontendDist))
	// 		return
	// 	}

	// 	fileServer.ServeHTTP(c.Writer, c.Request)
	// })

	r.Run(":8080")

	// os.Exit(0)
}
