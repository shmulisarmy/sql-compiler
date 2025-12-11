import { createMutable } from 'solid-js/store';

function reactive_mutable_set(obj: any, receiver: {}) {
  for (const key of Object.keys(obj)) {
    receiver[key as keyof typeof receiver] = obj[key] as keyof typeof receiver;
  }
}

type RemoteUpdate = {
  Type: "add" | "remove" | "update" | "load";
  Data: any;
  Path: string;
  Source_name: string;
};

function syncMessagesInto(receiver: {}) {
  return function (event: MessageEvent) {
    const update: RemoteUpdate = JSON.parse(event.data);
    console.log(update);

    switch (update.Type) {
      case "add":
        console.log("add", update.Data);
        let current = receiver;
        const all_expect_first_and_last = update.Path.split("/").slice(1, -1);
        const last_key = update.Path.split("/").pop();
        console.log({ all_expect_first_and_last });

        for (const key of all_expect_first_and_last) {
          if (!current[key as keyof typeof current]) {
            current[key as keyof typeof current] = {} as never;
          }
          current = current[key as keyof typeof current];
        }

        console.log({ current });
        console.log({ data: update.Data });
        console.log({ last_key });
        current[last_key as keyof typeof current] = JSON.parse(update.Data) as never;
        break;

      case "remove":
        console.log("remove", update.Data);
        let currentRemove = receiver;
        const pathPartsRemove = update.Path.split("/").slice(1);
        const lastKeyRemove = pathPartsRemove.pop();

        for (const key of pathPartsRemove) {
          if (!currentRemove[key as keyof typeof currentRemove]) {
            return; // Path doesn't exist, nothing to remove
          }
          currentRemove = currentRemove[key as keyof typeof currentRemove];
        }

        if (lastKeyRemove) {
          delete currentRemove[lastKeyRemove as keyof typeof currentRemove];
        }
        break;

      case "update":
        console.log("update", update.Data);
        let currentUpdate = receiver;
        const pathPartsUpdate = update.Path.split("/").slice(1);
        const lastKeyUpdate = pathPartsUpdate.pop();

        for (const key of pathPartsUpdate) {
          if (!currentUpdate[key as keyof typeof currentUpdate]) {
            console.error("Path does not exist for update:", update.Path);
            return;
          }
          currentUpdate = currentUpdate[key as keyof typeof currentUpdate];
        }

        if (lastKeyUpdate) {
          currentUpdate[lastKeyUpdate as keyof typeof currentUpdate] = JSON.parse(update.Data) as never;
        }
        break;

      case "load":
        reactive_mutable_set(JSON.parse(update.Data), receiver);
        break;

      default:
        console.log("unknown type", update);
        console.log("unknown type");
        break;
    }
  };
}

export function live_db(streaming_source: string): {} {
  const treeShapedReceiver = createMutable({});
  const ws = new WebSocket(streaming_source);
  ws.onmessage = syncMessagesInto(treeShapedReceiver);
  return treeShapedReceiver;
}
