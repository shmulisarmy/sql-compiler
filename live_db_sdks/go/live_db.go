package live_db

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type SyncType string

const (
	SyncTypeAdd    SyncType = "add"
	SyncTypeRemove SyncType = "remove"
	SyncTypeUpdate SyncType = "update"
	SyncTypeLoad   SyncType = "load"
)

type RemoteUpdate struct {
	Type       SyncType `json:"Type"`
	Data       string   `json:"Data"`
	Path       string   `json:"Path"`
	SourceName string   `json:"Source_name"`
	Timestamp  int64    `json:"Timestamp"`
}

type LiveDB struct {
	data    map[string]any
	mu      sync.RWMutex
	conn    *websocket.Conn
	onError func(error)
}

// NewLiveDB creates a new LiveDB instance and connects to the WebSocket streaming source
func NewLiveDB(streamingSource string) (*LiveDB, error) {
	db := &LiveDB{
		data:    make(map[string]any),
		onError: func(err error) { log.Printf("LiveDB error: %v", err) },
	}

	conn, _, err := websocket.DefaultDialer.Dial(streamingSource, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to websocket: %w", err)
	}
	db.conn = conn

	go db.listen()

	return db, nil
}

// SetErrorHandler sets a custom error handler
func (db *LiveDB) SetErrorHandler(handler func(error)) {
	db.onError = handler
}

// GetData returns a read-only copy of the current data
func (db *LiveDB) GetData() map[string]any {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.copyData(db.data)
}

// GetAtPath returns the value at a specific path
func (db *LiveDB) GetAtPath(path string) (any, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	parts := strings.Split(path, "/")
	parts = parts[1:] // Remove empty first element

	current := db.data
	for i, key := range parts {
		if i == len(parts)-1 {
			val, exists := current[key]
			return val, exists
		}

		next, exists := current[key]
		if !exists {
			return nil, false
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			return nil, false
		}
		current = nextMap
	}

	return nil, false
}

func (db *LiveDB) listen() {
	for {
		_, message, err := db.conn.ReadMessage()
		if err != nil {
			db.onError(fmt.Errorf("read error: %w", err))
			return
		}

		var update RemoteUpdate
		if err := json.Unmarshal(message, &update); err != nil {
			db.onError(fmt.Errorf("failed to unmarshal message: %w", err))
			continue
		}

		if err := db.handleUpdate(update); err != nil {
			db.onError(fmt.Errorf("failed to handle update: %w", err))
		}
	}
}

func (db *LiveDB) handleUpdate(update RemoteUpdate) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	switch update.Type {
	case SyncTypeAdd:
		return db.handleAdd(update)
	case SyncTypeRemove:
		return db.handleRemove(update)
	case SyncTypeUpdate:
		return db.handleUpdateData(update)
	case SyncTypeLoad:
		return db.handleLoad(update)
	default:
		return fmt.Errorf("unknown sync type: %s", update.Type)
	}
}

func (db *LiveDB) handleAdd(update RemoteUpdate) error {
	var data any
	if err := json.Unmarshal([]byte(update.Data), &data); err != nil {
		return fmt.Errorf("failed to unmarshal add data: %w", err)
	}

	parts := strings.Split(update.Path, "/")
	parts = parts[1:] // Remove empty first element from leading /

	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	current := db.data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]

		next, exists := current[key]
		if !exists {
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
			continue
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("path element %s is not a map", key)
		}
		current = nextMap
	}

	lastKey := parts[len(parts)-1]
	current[lastKey] = data

	return nil
}

func (db *LiveDB) handleRemove(update RemoteUpdate) error {
	parts := strings.Split(update.Path, "/")
	parts = parts[1:] // Remove empty first element

	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	current := db.data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]

		next, exists := current[key]
		if !exists {
			return nil // Path doesn't exist, nothing to remove
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("path element %s is not a map", key)
		}
		current = nextMap
	}

	lastKey := parts[len(parts)-1]
	delete(current, lastKey)

	return nil
}

func (db *LiveDB) handleUpdateData(update RemoteUpdate) error {
	var data any
	if err := json.Unmarshal([]byte(update.Data), &data); err != nil {
		return fmt.Errorf("failed to unmarshal update data: %w", err)
	}

	parts := strings.Split(update.Path, "/")
	parts = parts[1:] // Remove empty first element

	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	current := db.data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]

		next, exists := current[key]
		if !exists {
			return fmt.Errorf("path does not exist for update: %s", update.Path)
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("path element %s is not a map", key)
		}
		current = nextMap
	}

	lastKey := parts[len(parts)-1]
	current[lastKey] = data

	return nil
}

func (db *LiveDB) handleLoad(update RemoteUpdate) error {
	var data map[string]any
	if err := json.Unmarshal([]byte(update.Data), &data); err != nil {
		return fmt.Errorf("failed to unmarshal load data: %w", err)
	}

	// Replace entire data structure
	db.data = data

	return nil
}

func (db *LiveDB) copyData(data map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range data {
		switch val := v.(type) {
		case map[string]any:
			result[k] = db.copyData(val)
		default:
			result[k] = val
		}
	}
	return result
}

// Close closes the WebSocket connection
func (db *LiveDB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}
