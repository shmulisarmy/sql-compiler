package local_live_db

import (
	"encoding/json"
	"fmt"
	"strings"

	eventEmitterTree "sql-compiler/eventEmitterTree"
)

type LocalLiveDB struct {
	Data map[string]any
}

func (db *LocalLiveDB) HandleUpdate(update eventEmitterTree.SyncMessage) error {

	switch update.Type {
	case eventEmitterTree.SyncTypeAdd:
		return db.handleAdd(update)
	case eventEmitterTree.SyncTypeRemove:
		return db.handleRemove(update)
	case eventEmitterTree.SyncTypeUpdate:
		return db.handleUpdateData(update)
	case eventEmitterTree.LoadInitialData:
		return db.handleLoad(update)
	default:
		return fmt.Errorf("unknown sync type: %s", update.Type)
	}
}

func (db *LocalLiveDB) handleAdd(update eventEmitterTree.SyncMessage) error {
	var data any
	if err := json.Unmarshal([]byte(update.Data), &data); err != nil {
		return fmt.Errorf("failed to unmarshal add data: %w", err)
	}

	parts := strings.Split(update.Path, "/")
	parts = parts[1:] // Remove empty first element from leading /

	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	current := db.Data
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

func (db *LocalLiveDB) handleRemove(update eventEmitterTree.SyncMessage) error {
	parts := strings.Split(update.Path, "/")
	parts = parts[1:] // Remove empty first element

	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	current := db.Data
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

func (db *LocalLiveDB) handleUpdateData(update eventEmitterTree.SyncMessage) error {
	var data any
	if err := json.Unmarshal([]byte(update.Data), &data); err != nil {
		return fmt.Errorf("failed to unmarshal update data: %w", err)
	}

	parts := strings.Split(update.Path, "/")
	parts = parts[1:] // Remove empty first element

	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	current := db.Data
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

func (db *LocalLiveDB) handleLoad(update eventEmitterTree.SyncMessage) error {
	var data map[string]any
	if err := json.Unmarshal([]byte(update.Data), &data); err != nil {
		return fmt.Errorf("failed to unmarshal load data: %w", err)
	}

	// Replace entire data structure
	db.Data = data

	return nil
}

func (db *LocalLiveDB) copyData(data map[string]any) map[string]any {
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
