// The client of the WebView bridge, which the Authgear SDKs provide in their own WebView.
// See docs/specs/webview-bridge.md.

export interface WebViewBridgeReply {
  status: string;
  [key: string]: unknown;
}

export interface WebViewBridge {
  // Sends {id, type, ...fields} and resolves with the reply object.
  // Rejects when the transport fails or the reply is not a JSON object.
  request(
    type: string,
    fields?: Record<string, unknown>
  ): Promise<WebViewBridgeReply>;
}

interface IOSMessageHandler {
  postMessage(text: string): Promise<unknown>;
}

interface AndroidBridge {
  postMessage(text: string): void;
  onmessage: ((event: MessageEvent) => void) | null;
}

let nextID = 1;

function parseReply(text: unknown): WebViewBridgeReply {
  if (typeof text !== "string") {
    throw new Error("webview bridge: the reply is not a string");
  }
  const reply: unknown = JSON.parse(text);
  if (reply == null || typeof reply !== "object" || Array.isArray(reply)) {
    throw new Error("webview bridge: the reply is not a JSON object");
  }
  return reply as WebViewBridgeReply;
}

function makeBridge(
  send: (id: string, text: string) => Promise<unknown>
): WebViewBridge {
  return {
    async request(type, fields = {}) {
      const id = String(nextID++);
      const text = JSON.stringify({ ...fields, id, type });
      return parseReply(await send(id, text));
    },
  };
}

function getIOSMessageHandler(): IOSMessageHandler | null {
  const handler = (window as any).webkit?.messageHandlers?.authgearBridge;
  if (handler == null || typeof handler.postMessage !== "function") {
    return null;
  }
  return handler;
}

function getAndroidBridge(): AndroidBridge | null {
  const bridge = (window as any).authgearBridge;
  if (bridge == null || typeof bridge.postMessage !== "function") {
    return null;
  }
  return bridge;
}

// The Android bridge delivers every reply to one onmessage,
// so it is set once for each bridge object.
let androidBridge: { source: AndroidBridge; bridge: WebViewBridge } | null =
  null;

function makeAndroidBridge(source: AndroidBridge): WebViewBridge {
  if (androidBridge?.source === source) {
    return androidBridge.bridge;
  }

  const pending = new Map<string, (text: string) => void>();
  source.onmessage = (event: MessageEvent) => {
    let id: unknown;
    try {
      id = JSON.parse(event.data)?.id;
    } catch (e: unknown) {
      console.warn("webview bridge: the reply is not JSON", e);
      return;
    }
    if (typeof id !== "string") {
      return;
    }
    const resolve = pending.get(id);
    if (resolve == null) {
      return;
    }
    pending.delete(id);
    resolve(event.data);
  };

  const bridge = makeBridge(
    async (id, text) =>
      new Promise<string>((resolve, reject) => {
        pending.set(id, resolve);
        try {
          source.postMessage(text);
        } catch (e: unknown) {
          pending.delete(id);
          reject(e);
        }
      })
  );
  androidBridge = { source, bridge };
  return bridge;
}

// Unlike getWebViewBridge, it does not take over the Android bridge's onmessage.
export function hasWebViewBridge(): boolean {
  return getIOSMessageHandler() != null || getAndroidBridge() != null;
}

// Returns null when the page is not in an Authgear SDK WebView with the bridge.
export function getWebViewBridge(): WebViewBridge | null {
  const ios = getIOSMessageHandler();
  if (ios != null) {
    return makeBridge(async (_id, text) => ios.postMessage(text));
  }
  const android = getAndroidBridge();
  if (android != null) {
    return makeAndroidBridge(android);
  }
  return null;
}

// Resolves with the reply's capabilities, or [] when the request fails
// (iOS rejects a page whose origin is not allowed) or the reply is not ok.
export async function getWebViewBridgeCapabilities(
  bridge: WebViewBridge
): Promise<string[]> {
  let reply: WebViewBridgeReply;
  try {
    reply = await bridge.request("bridge.capabilities");
  } catch (e: unknown) {
    console.warn("webview bridge: bridge.capabilities failed", e);
    return [];
  }
  if (reply.status !== "ok" || !Array.isArray(reply.capabilities)) {
    return [];
  }
  return reply.capabilities.filter((c): c is string => typeof c === "string");
}
