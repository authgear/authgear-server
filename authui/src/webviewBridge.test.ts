import { jest, describe, it, expect, afterEach } from "@jest/globals";
import {
  getWebViewBridge,
  getWebViewBridgeCapabilities,
  hasWebViewBridge,
} from "./webviewBridge";

afterEach(() => {
  delete (window as any).webkit;
  delete (window as any).authgearBridge;
});

function installIOS(
  postMessage: (text: string) => Promise<unknown>
): jest.Mock<(text: string) => Promise<unknown>> {
  const mock = jest.fn(postMessage);
  (window as any).webkit = {
    messageHandlers: { authgearBridge: { postMessage: mock } },
  };
  return mock;
}

interface FakeAndroidBridge {
  postMessage: jest.Mock<(text: string) => void>;
  onmessage: ((event: MessageEvent) => void) | null;
  reply(reply: unknown): void;
}

function installAndroid(): FakeAndroidBridge {
  const bridge: FakeAndroidBridge = {
    postMessage: jest.fn(() => {}),
    onmessage: null,
    reply(reply: unknown) {
      bridge.onmessage?.(
        new MessageEvent("message", { data: JSON.stringify(reply) })
      );
    },
  };
  (window as any).authgearBridge = bridge;
  return bridge;
}

function sentMessage(mock: jest.Mock<(text: string) => any>, i: number): any {
  const text = mock.mock.calls[i][0];
  expect(typeof text).toEqual("string");
  return JSON.parse(text);
}

async function flush(): Promise<void> {
  for (let i = 0; i < 10; i++) {
    await Promise.resolve();
  }
}

describe("getWebViewBridge", () => {
  it("returns null when there is no bridge", () => {
    expect(getWebViewBridge()).toBeNull();
  });

  it("sends a message as text and resolves with the reply on iOS", async () => {
    const postMessage = installIOS(async (text) =>
      JSON.stringify({ id: JSON.parse(text).id, status: "ok", value: 42 })
    );

    const bridge = getWebViewBridge();
    expect(bridge).not.toBeNull();
    const reply = await bridge?.request("test.echo", { foo: "bar" });

    expect(postMessage).toHaveBeenCalledTimes(1);
    const message = sentMessage(postMessage, 0);
    expect(message).toEqual({
      id: expect.any(String),
      type: "test.echo",
      foo: "bar",
    });
    expect(reply).toEqual({ id: message.id, status: "ok", value: 42 });
  });

  it("uses unique ids", async () => {
    const postMessage = installIOS(async (text) =>
      JSON.stringify({ id: JSON.parse(text).id, status: "ok" })
    );

    const bridge = getWebViewBridge();
    await bridge?.request("test.echo");
    await getWebViewBridge()?.request("test.echo");

    expect(sentMessage(postMessage, 0).id).not.toEqual(
      sentMessage(postMessage, 1).id
    );
  });

  it("resolves each request with its own reply on Android", async () => {
    const android = installAndroid();
    const bridge = getWebViewBridge();
    expect(bridge).not.toBeNull();

    const first = bridge?.request("test.first");
    const second = bridge?.request("test.second");
    expect(android.postMessage).toHaveBeenCalledTimes(2);
    const firstID = sentMessage(android.postMessage, 0).id;
    const secondID = sentMessage(android.postMessage, 1).id;
    expect(sentMessage(android.postMessage, 0).type).toEqual("test.first");
    expect(sentMessage(android.postMessage, 1).type).toEqual("test.second");

    android.reply({ id: secondID, status: "ok", which: "second" });
    android.reply({ id: firstID, status: "ok", which: "first" });

    await expect(first).resolves.toMatchObject({ which: "first" });
    await expect(second).resolves.toMatchObject({ which: "second" });
  });

  it("ignores a reply with an unknown id on Android", async () => {
    const android = installAndroid();
    const bridge = getWebViewBridge();

    let settled = false;
    const promise = bridge?.request("test.echo");
    void promise?.then(
      () => (settled = true),
      () => (settled = true)
    );
    const id = sentMessage(android.postMessage, 0).id;

    const warn = jest.spyOn(console, "warn").mockImplementation(() => {});
    android.reply({ id: "unknown", status: "ok" });
    android.onmessage?.(new MessageEvent("message", { data: "not json" }));
    await flush();
    expect(settled).toEqual(false);
    expect(warn).toHaveBeenCalledTimes(1);
    warn.mockRestore();

    android.reply({ id, status: "ok", which: "mine" });
    await expect(promise).resolves.toMatchObject({ which: "mine" });
  });

  it("sets onmessage once on Android", async () => {
    const android = installAndroid();
    getWebViewBridge();
    const onmessage = android.onmessage;
    expect(onmessage).not.toBeNull();

    const pending = getWebViewBridge()?.request("test.echo");
    expect(android.onmessage).toBe(onmessage);
    android.reply({
      id: sentMessage(android.postMessage, 0).id,
      status: "ok",
    });
    await expect(pending).resolves.toMatchObject({ status: "ok" });
  });

  it("prefers iOS when both exist", async () => {
    const postMessage = installIOS(async (text) =>
      JSON.stringify({ id: JSON.parse(text).id, status: "ok" })
    );
    const android = installAndroid();

    await getWebViewBridge()?.request("test.echo");

    expect(postMessage).toHaveBeenCalledTimes(1);
    expect(android.postMessage).not.toHaveBeenCalled();
  });

  it("rejects a reply that is not a JSON object", async () => {
    for (const text of ["not json", "[]", "null", "42", '"ok"']) {
      installIOS(async () => text);
      await expect(getWebViewBridge()?.request("test.echo")).rejects.toThrow();
    }
  });

  it("rejects when the transport fails", async () => {
    installIOS(async () => {
      throw new Error("origin is not allowed");
    });
    await expect(getWebViewBridge()?.request("test.echo")).rejects.toThrow(
      "origin is not allowed"
    );
  });
});

describe("hasWebViewBridge", () => {
  it("returns false when there is no bridge", () => {
    expect(hasWebViewBridge()).toEqual(false);
  });

  it("returns true on iOS", () => {
    installIOS(async () => "{}");
    expect(hasWebViewBridge()).toEqual(true);
  });

  it("returns true on Android without setting onmessage", () => {
    const android = installAndroid();
    expect(hasWebViewBridge()).toEqual(true);
    expect(android.onmessage).toBeNull();
  });
});

describe("getWebViewBridgeCapabilities", () => {
  function installCapabilitiesReply(reply: Record<string, unknown>) {
    return installIOS(async (text) =>
      JSON.stringify({ id: JSON.parse(text).id, ...reply })
    );
  }

  it("returns the capabilities", async () => {
    const postMessage = installCapabilitiesReply({
      status: "ok",
      capabilities: ["credentials.get", "credentials.get:publicKey"],
    });

    const bridge = getWebViewBridge();
    if (bridge == null) {
      throw new Error("no bridge");
    }
    await expect(getWebViewBridgeCapabilities(bridge)).resolves.toEqual([
      "credentials.get",
      "credentials.get:publicKey",
    ]);
    expect(sentMessage(postMessage, 0).type).toEqual("bridge.capabilities");
  });

  it("returns [] and warns when the request is rejected", async () => {
    const warn = jest.spyOn(console, "warn").mockImplementation(() => {});
    installIOS(async () => {
      throw new Error("origin is not allowed");
    });
    const bridge = getWebViewBridge();
    if (bridge == null) {
      throw new Error("no bridge");
    }
    await expect(getWebViewBridgeCapabilities(bridge)).resolves.toEqual([]);
    expect(warn).toHaveBeenCalled();
    warn.mockRestore();
  });

  it("returns [] when the reply is not ok", async () => {
    installCapabilitiesReply({
      status: "unsupported",
      capabilities: ["credentials.get"],
    });
    const bridge = getWebViewBridge();
    if (bridge == null) {
      throw new Error("no bridge");
    }
    await expect(getWebViewBridgeCapabilities(bridge)).resolves.toEqual([]);
  });

  it("returns [] when capabilities is not an array", async () => {
    installCapabilitiesReply({
      status: "ok",
      capabilities: "credentials.get",
    });
    const bridge = getWebViewBridge();
    if (bridge == null) {
      throw new Error("no bridge");
    }
    await expect(getWebViewBridgeCapabilities(bridge)).resolves.toEqual([]);
  });
});
