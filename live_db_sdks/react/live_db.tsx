import { useState, useEffect, useRef } from 'react';

type RemoteUpdate = {
  Type: "add" | "remove" | "update" | "load";
  Data: any;
  Path: string;
  Source_name: string;
};

function applyUpdate(state: any, update: RemoteUpdate): any {
  const newState = JSON.parse(JSON.stringify(state)); // Deep clone

  switch (update.Type) {
    case "add": {
      let current = newState;
      const pathParts = update.Path.split("/").slice(1, -1);
      const lastKey = update.Path.split("/").pop();

      for (const key of pathParts) {
        if (!current[key]) {
          current[key] = {};
        }
        current = current[key];
      }

      if (lastKey) {
        current[lastKey] = JSON.parse(update.Data);
      }
      break;
    }

    case "remove": {
      let current = newState;
      const pathParts = update.Path.split("/").slice(1);
      const lastKey = pathParts.pop();

      for (const key of pathParts) {
        if (!current[key]) {
          return state; // Path doesn't exist, return unchanged
        }
        current = current[key];
      }

      if (lastKey) {
        delete current[lastKey];
      }
      break;
    }

    case "update": {
      let current = newState;
      const pathParts = update.Path.split("/").slice(1);
      const lastKey = pathParts.pop();

      for (const key of pathParts) {
        if (!current[key]) {
          console.error("Path does not exist for update:", update.Path);
          return state;
        }
        current = current[key];
      }

      if (lastKey) {
        current[lastKey] = JSON.parse(update.Data);
      }
      break;
    }

    case "load": {
      return JSON.parse(update.Data);
    }

    default:
      console.log("unknown type", update);
      return state;
  }

  return newState;
}

export function useLiveDB(streamingSource: string) {
  const [data, setData] = useState<any>({});
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const ws = new WebSocket(streamingSource);
    wsRef.current = ws;

    ws.onmessage = (event: MessageEvent) => {
      const update: RemoteUpdate = JSON.parse(event.data);
      console.log(update);

      setData((prevData: any) => applyUpdate(prevData, update));
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    ws.onclose = () => {
      console.log("WebSocket connection closed");
    };

    return () => {
      ws.close();
    };
  }, [streamingSource]);

  return data;
}

// Alternative: Non-hook version for class components or custom usage
export class LiveDB {
  private data: any = {};
  private ws: WebSocket | null = null;
  private listeners: Set<(data: any) => void> = new Set();

  constructor(streamingSource: string) {
    this.ws = new WebSocket(streamingSource);

    this.ws.onmessage = (event: MessageEvent) => {
      const update: RemoteUpdate = JSON.parse(event.data);
      console.log(update);

      this.data = applyUpdate(this.data, update);
      this.notifyListeners();
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    this.ws.onclose = () => {
      console.log("WebSocket connection closed");
    };
  }

  subscribe(listener: (data: any) => void) {
    this.listeners.add(listener);
    listener(this.data); // Emit current state immediately

    return () => {
      this.listeners.delete(listener);
    };
  }

  private notifyListeners() {
    this.listeners.forEach((listener) => listener(this.data));
  }

  getData() {
    return this.data;
  }

  close() {
    if (this.ws) {
      this.ws.close();
    }
  }
}
