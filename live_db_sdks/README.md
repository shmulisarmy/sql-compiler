# Live DB SDKs

Client SDKs for real-time database synchronization with SQL Compiler's EventEmitterTree system.

## Overview

The Live DB SDKs provide client-side implementations that sync with the backend's `eventEmitterTree` package. The backend emits hierarchical sync messages over WebSocket, and these SDKs reconstruct the data structure on the client side with framework-specific reactivity.

## Architecture

### Backend (Go - eventEmitterTree)

The backend `eventEmitterTree/eventEmitterTree.go` subscribes to observables (SQL queries with nested subqueries) and emits sync messages:

```go
type SyncMessage struct {
    Type      SyncType  // "add", "remove", "update", "load"
    Data      string    // JSON-serialized row data
    Path      string    // Hierarchical path like "/primary_key/field/nested_key"
    Timestamp int64
}
```

### Path Structure

Paths represent the hierarchical structure of nested data:
- `/user_123` - Top-level user
- `/user_123/posts/post_456` - Nested post under user
- `/user_123/friends/user_789/posts/post_101` - Deeply nested structure

For `GroupBy` operations, the group value is included in the path:
- `/groupValue/user_123` - Item in a group

## Available SDKs

### 1. Go SDK (`/go`)

For Go applications that need to consume live data streams.

```go
import "sql-compiler/live_db_sdks/go"

db, err := live_db.NewLiveDB("ws://localhost:8080/stream-data")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Get current data
data := db.GetData()

// Get value at specific path
value, exists := db.GetAtPath("/user_123/posts")
```

**Features:**
- Thread-safe with `sync.RWMutex`
- Returns defensive copies to prevent external mutations
- Custom error handlers
- Path-based access methods

### 2. Solid.js SDK (`/solid`)

For Solid.js applications with fine-grained reactivity.

```tsx
import { live_db } from '@sql-compiler/live-db-solid';

function App() {
  const data = live_db('ws://localhost:8080/stream-data');

  return (
    <For each={Object.entries(data)}>
      {([key, value]) => <div>{key}: {JSON.stringify(value)}</div>}
    </For>
  );
}
```

**Features:**
- Uses Solid's `createMutable` for fine-grained reactivity
- Automatic updates without re-renders
- Direct object mutation with reactive tracking
- Complete implementation of add, remove, update, load operations

### 3. React SDK (`/react`)

For React applications with hooks or class components.

#### Hook API (Recommended)

```tsx
import { useLiveDB } from '@sql-compiler/live-db-react';

function App() {
  const data = useLiveDB('ws://localhost:8080/stream-data');

  return (
    <div>
      {Object.entries(data).map(([key, value]) => (
        <div key={key}>{key}: {JSON.stringify(value)}</div>
      ))}
    </div>
  );
}
```

#### Class API

```tsx
import { LiveDB } from '@sql-compiler/live-db-react';

class App extends Component {
  liveDB = new LiveDB('ws://localhost:8080/stream-data');

  componentDidMount() {
    this.liveDB.subscribe((data) => this.setState({ data }));
  }

  componentWillUnmount() {
    this.liveDB.close();
  }
}
```

**Features:**
- Modern hooks API and class-based API
- Immutable state updates for React's reconciliation
- Deep cloning for proper change detection
- Subscribe/unsubscribe pattern

## Sync Operations

All SDKs handle these operations:

### Add
Adds a new item at the specified path, creating intermediate objects as needed.

```json
{
  "Type": "add",
  "Path": "/user_123/posts/post_456",
  "Data": "{\"title\":\"Hello\",\"body\":\"World\"}"
}
```

### Remove
Removes an item at the specified path.

```json
{
  "Type": "remove",
  "Path": "/user_123/posts/post_456",
  "Data": "{...}"
}
```

### Update
Updates an existing item at the specified path.

```json
{
  "Type": "update",
  "Path": "/user_123/posts/post_456",
  "Data": "{\"title\":\"Updated\",\"body\":\"Content\"}"
}
```

### Load
Replaces the entire data structure (used for initial load).

```json
{
  "Type": "load",
  "Path": "/",
  "Data": "{\"user_123\":{...},\"user_456\":{...}}"
}
```

## Testing

The system includes integration tests that verify data consistency across multiple clients:

```tsx
// See frontend/src/integration_tests.tsx
const client1 = live_db('ws://localhost:8080/stream-data');
// ... trigger updates ...
const client2 = live_db('ws://localhost:8080/stream-data');
// Both clients should have identical data regardless of connection timing
```

## Directory Structure

```
live_db_sdks/
├── README.md              # This file
├── go/                    # Go SDK
│   ├── live_db.go
│   └── go.mod
├── solid/                 # Solid.js SDK
│   ├── live_db.tsx
│   ├── package.json
│   └── README.md
└── react/                 # React SDK
    ├── live_db.tsx
    ├── package.json
    └── README.md
```

## Development

Each SDK is designed to be framework-idiomatic:
- **Go**: Uses channels, goroutines, and mutex locks
- **Solid**: Uses `createMutable` for fine-grained reactivity
- **React**: Uses `useState` with immutable updates

All SDKs maintain the same core logic for path traversal and data synchronization while adapting to their framework's patterns.
