# Live DB - Solid.js SDK

Solid.js SDK for real-time database synchronization using WebSockets.

## Installation

```bash
npm install solid-js
```

## Usage

```tsx
import { live_db } from './live_db';

function App() {
  const data = live_db('ws://localhost:8080/stream-data');

  return (
    <div>
      <For each={Object.entries(data)}>
        {([key, value]) => (
          <div>{key}: {JSON.stringify(value)}</div>
        )}
      </For>
    </div>
  );
}
```

## Features

- Automatic WebSocket connection management
- Reactive updates using Solid's `createMutable`
- Supports nested data structures
- Handles add, remove, update, and load operations
