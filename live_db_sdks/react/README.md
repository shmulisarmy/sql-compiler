# Live DB - React SDK

React SDK for real-time database synchronization using WebSockets.

## Installation

```bash
npm install react
```

## Usage

### Hook API (Recommended)

```tsx
import { useLiveDB } from './live_db';

function App() {
  const data = useLiveDB('ws://localhost:8080/stream-data');

  return (
    <div>
      {Object.entries(data).map(([key, value]) => (
        <div key={key}>
          {key}: {JSON.stringify(value)}
        </div>
      ))}
    </div>
  );
}
```

### Class API (For class components or custom usage)

```tsx
import { LiveDB } from './live_db';
import { Component } from 'react';

class App extends Component {
  liveDB: LiveDB;
  state = { data: {} };

  componentDidMount() {
    this.liveDB = new LiveDB('ws://localhost:8080/stream-data');
    this.liveDB.subscribe((data) => {
      this.setState({ data });
    });
  }

  componentWillUnmount() {
    this.liveDB.close();
  }

  render() {
    return (
      <div>
        {Object.entries(this.state.data).map(([key, value]) => (
          <div key={key}>
            {key}: {JSON.stringify(value)}
          </div>
        ))}
      </div>
    );
  }
}
```

## Features

- React hooks API (`useLiveDB`)
- Class-based API (`LiveDB`)
- Automatic WebSocket connection management
- Immutable state updates for optimal React rendering
- Supports nested data structures
- Handles add, remove, update, and load operations

## API

### `useLiveDB(streamingSource: string)`

A React hook that connects to a WebSocket and returns the synchronized data.

### `LiveDB`

A class-based implementation that can be used with class components or custom logic.

- `constructor(streamingSource: string)` - Creates connection
- `subscribe(listener: (data: any) => void)` - Subscribe to data updates
- `getData()` - Get current data snapshot
- `close()` - Close WebSocket connection
