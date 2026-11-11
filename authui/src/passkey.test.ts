import {
  jest,
  describe,
  it,
  expect,
  beforeAll,
  afterAll,
  beforeEach,
  afterEach,
} from "@jest/globals";
import { Application } from "@hotwired/stimulus";
import axios from "axios";
import {
  AuthflowPasskeyRequestController,
  PasskeyAutofillController,
} from "./passkey";

jest.mock("axios");

const axiosMock = jest.mocked(axios);

// "AQID" is [1, 2, 3] in base64url.
const OPTIONS_JSON = JSON.stringify({
  publicKey: {
    challenge: "AQID",
    timeout: 3600000,
    rpId: "localhost",
    userVerification: "preferred",
    extensions: { uvm: true },
  },
});

interface PendingRequest {
  options: CredentialRequestOptions;
  resolve: (credential: unknown) => void;
  reject: (err: unknown) => void;
  settled: boolean;
  // The number of other requests still pending when this one was made.
  othersPending: number;
}

let application: Application;
let requests: PendingRequest[];
let isConditionalMediationAvailable: jest.Mock<() => Promise<boolean>>;
let submitClicks: number;

beforeAll(() => {
  jest.useFakeTimers();
  application = Application.start();
  application.register("passkey-autofill", PasskeyAutofillController);
  application.register(
    "authflow-passkey-request",
    AuthflowPasskeyRequestController
  );
});

afterAll(() => {
  application.stop();
  jest.useRealTimers();
});

beforeEach(() => {
  requests = [];
  submitClicks = 0;
  isConditionalMediationAvailable = jest.fn(async () => true);
  Object.defineProperty(window, "PublicKeyCredential", {
    configurable: true,
    value: class {
      static isConditionalMediationAvailable = isConditionalMediationAvailable;
    },
  });
  Object.defineProperty(window.navigator, "credentials", {
    configurable: true,
    value: {
      get: jest.fn(
        async (options: CredentialRequestOptions) =>
          new Promise((resolve, reject) => {
            const request: PendingRequest = {
              options,
              resolve: (credential) => {
                request.settled = true;
                resolve(credential);
              },
              reject: (err) => {
                request.settled = true;
                reject(err);
              },
              settled: false,
              othersPending: requests.filter((r) => !r.settled).length,
            };
            requests.push(request);
            options.signal?.addEventListener("abort", () => {
              request.reject(
                new DOMException("The operation was aborted.", "AbortError")
              );
            });
          })
      ),
    },
  });
  jest.spyOn(console, "error").mockImplementation(() => {});
  jest.spyOn(console, "warn").mockImplementation(() => {});
});

afterEach(async () => {
  document.body.innerHTML = "";
  await flush();
  jest.clearAllTimers();
  jest.restoreAllMocks();
  axiosMock.mockReset();
  delete (window as any).webkit;
  delete (window as any).authgearBridge;
  window.sessionStorage.clear();
});

// Stimulus connects controllers from a MutationObserver,
// and the controller awaits a few promises before calling navigator.credentials.get.
async function flush(): Promise<void> {
  for (let i = 0; i < 10; i++) {
    await Promise.resolve();
  }
}

async function render(): Promise<void> {
  document.body.innerHTML = `
    <div data-controller="passkey-autofill" data-passkey-autofill-options-value='${OPTIONS_JSON}'>
      <input type="hidden" data-passkey-autofill-target="input">
      <button type="button" data-passkey-autofill-target="submit"></button>
    </div>
    <div data-controller="authflow-passkey-request" data-authflow-passkey-request-options-value='${OPTIONS_JSON}'>
      <button type="button" id="use-passkey" data-action="click->authflow-passkey-request#use" data-authflow-passkey-request-target="button"></button>
      <input type="hidden" data-authflow-passkey-request-target="input">
      <button type="button" data-authflow-passkey-request-target="submit"></button>
    </div>
  `;
  document
    .querySelector("button")
    ?.addEventListener("click", () => (submitClicks += 1));
  await flush();
}

async function advance(ms: number): Promise<void> {
  jest.advanceTimersByTime(ms);
  await flush();
}

function makeCredential() {
  return {
    id: "credential-id",
    type: "public-key",
    response: {
      authenticatorData: new Uint8Array([1]).buffer,
      clientDataJSON: new Uint8Array([2]).buffer,
      signature: new Uint8Array([3]).buffer,
      userHandle: null,
    },
    getClientExtensionResults: () => ({}),
  };
}

async function showPage(persisted: boolean): Promise<void> {
  window.dispatchEvent(new PageTransitionEvent("pageshow", { persisted }));
  await flush();
}

type BridgeHandler = (
  message: Record<string, any>
) => Promise<Record<string, unknown>>;

// The messages sent through the fake bridge, parsed.
let bridgeMessages: Record<string, any>[];

// Installs a fake WebView bridge on window, in the shape of the platform.
// handler answers each message with the reply, without the id.
function installBridge(platform: "ios" | "android", handler: BridgeHandler) {
  bridgeMessages = [];
  const answer = async (text: string): Promise<string> => {
    const message = JSON.parse(text);
    bridgeMessages.push(message);
    const reply = await handler(message);
    return JSON.stringify({ id: message.id, ...reply });
  };
  if (platform === "ios") {
    (window as any).webkit = {
      messageHandlers: { authgearBridge: { postMessage: answer } },
    };
    return;
  }
  const android: {
    postMessage: (text: string) => void;
    onmessage: ((event: MessageEvent) => void) | null;
  } = {
    postMessage: (text) => {
      void answer(text).then(
        (data) => android.onmessage?.(new MessageEvent("message", { data })),
        () => {}
      );
    },
    onmessage: null,
  };
  (window as any).authgearBridge = android;
}

function notAllowed(): DOMException {
  return new DOMException("The operation is not allowed.", "NotAllowedError");
}

describe("PasskeyAutofillController", () => {
  it("requests with the options on the page and conditional mediation", async () => {
    await render();

    expect(requests).toHaveLength(1);
    const options = requests[0].options;
    expect(options.mediation).toEqual("conditional");
    expect(options.signal).toBeDefined();
    expect(Array.from(options.publicKey?.challenge as Uint8Array)).toEqual([
      1, 2, 3,
    ]);
    expect(options.publicKey?.rpId).toEqual("localhost");
    expect(axiosMock).not.toHaveBeenCalled();
  });

  it("submits the chosen passkey", async () => {
    await render();

    requests[0].resolve(makeCredential());
    await flush();

    const input = document.querySelector("input");
    expect(JSON.parse(input?.value ?? "")).toMatchObject({
      id: "credential-id",
      type: "public-key",
    });
    expect(submitClicks).toEqual(1);

    // The page changes, so there is no retry.
    await advance(60000);
    expect(requests).toHaveLength(1);
  });

  it("retries with the same options after the request fails", async () => {
    await render();

    await advance(10000);
    requests[0].reject(notAllowed());
    await flush();
    expect(requests).toHaveLength(1);

    await advance(1000);
    expect(requests).toHaveLength(2);
    expect(requests[1].options.mediation).toEqual("conditional");
    expect(
      Array.from(requests[1].options.publicKey?.challenge as Uint8Array)
    ).toEqual([1, 2, 3]);
    expect(axiosMock).not.toHaveBeenCalled();
  });

  it("stops retrying after the request fails instantly three times in a row", async () => {
    await render();

    requests[0].reject(notAllowed());
    await flush();
    await advance(1000);
    expect(requests).toHaveLength(2);

    requests[1].reject(notAllowed());
    await flush();
    await advance(1000);
    expect(requests).toHaveLength(3);

    requests[2].reject(notAllowed());
    await flush();
    await advance(60000);
    expect(requests).toHaveLength(3);
  });

  it("counts instant failures in a row only", async () => {
    await render();

    requests[0].reject(notAllowed());
    await flush();
    await advance(1000);
    requests[1].reject(notAllowed());
    await flush();
    await advance(1000);
    expect(requests).toHaveLength(3);

    // A request that was pending for a while resets the count.
    await advance(10000);
    requests[2].reject(notAllowed());
    await flush();
    await advance(1000);
    expect(requests).toHaveLength(4);

    requests[3].reject(notAllowed());
    await flush();
    await advance(1000);
    expect(requests).toHaveLength(5);
  });

  it("does not retry an error that a retry cannot fix", async () => {
    await render();

    requests[0].reject(
      new DOMException("The RP ID is not valid.", "SecurityError")
    );
    await flush();
    await advance(60000);

    expect(requests).toHaveLength(1);
  });

  it("aborts the request and stops retrying when disconnected", async () => {
    await render();
    const signal = requests[0].options.signal;

    document.body.innerHTML = "";
    await flush();

    expect(signal?.aborted).toEqual(true);
    await advance(60000);
    expect(requests).toHaveLength(1);
  });

  it("does not request when disconnected before the availability check resolves", async () => {
    let resolveAvailable: (available: boolean) => void = () => {};
    isConditionalMediationAvailable.mockImplementation(
      async () =>
        new Promise((resolve) => {
          resolveAvailable = resolve;
        })
    );
    await render();

    document.body.innerHTML = "";
    await flush();
    resolveAvailable(true);
    await advance(60000);

    expect(requests).toHaveLength(0);
  });

  it("does not request when conditional mediation is unavailable", async () => {
    isConditionalMediationAvailable.mockResolvedValue(false);
    await render();
    await advance(60000);

    expect(requests).toHaveLength(0);
  });

  it("gives each request its own abort signal", async () => {
    await render();
    await advance(10000);
    requests[0].reject(notAllowed());
    await flush();
    await advance(1000);
    expect(requests).toHaveLength(2);

    expect(requests[1].options.signal).not.toBe(requests[0].options.signal);

    // The first request ended, so nothing aborts its signal again.
    document.body.innerHTML = "";
    await flush();
    expect(requests[0].options.signal?.aborted).toEqual(false);
    expect(requests[1].options.signal?.aborted).toEqual(true);
  });

  it("aborts nothing when it stops while waiting to retry", async () => {
    await render();
    await advance(10000);
    requests[0].reject(notAllowed());
    await flush();

    document.body.innerHTML = "";
    await flush();
    await advance(60000);

    expect(requests).toHaveLength(1);
    expect(requests[0].options.signal?.aborted).toEqual(false);
  });

  it("starts again when the page is restored from the cache", async () => {
    await render();
    expect(requests).toHaveLength(1);

    await showPage(false);
    expect(requests).toHaveLength(1);

    await showPage(true);
    expect(requests).toHaveLength(2);
    expect(requests[0].options.signal?.aborted).toEqual(true);
    expect(requests[1].options.mediation).toEqual("conditional");
    expect(requests[1].othersPending).toEqual(0);
  });

  it("steps aside while the passkey dialog is open", async () => {
    await render();
    expect(requests).toHaveLength(1);

    document.getElementById("use-passkey")?.click();
    await flush();

    expect(requests).toHaveLength(2);
    expect(requests[0].options.signal?.aborted).toEqual(true);
    expect(requests[1].options.mediation).toBeUndefined();
    expect(requests[1].othersPending).toEqual(0);

    // Cancel the dialog. Autofill does not start again on this page.
    requests[1].reject(notAllowed());
    await flush();
    await advance(60000);
    expect(requests).toHaveLength(2);

    // The button opens its dialog again without cancelling an autofill request.
    document.getElementById("use-passkey")?.click();
    await flush();
    expect(requests).toHaveLength(3);
    expect(requests[2].options.mediation).toBeUndefined();
    expect(requests[2].othersPending).toEqual(0);

    // Autofill is offered again when the page is restored from the cache.
    requests[2].reject(notAllowed());
    await flush();
    await showPage(true);
    expect(requests).toHaveLength(4);
    expect(requests[3].options.mediation).toEqual("conditional");
  });
});

describe("AuthflowPasskeyRequestController", () => {
  it("disables the button while the request is pending", async () => {
    await render();
    const button = document.getElementById("use-passkey") as HTMLButtonElement;
    expect(button.disabled).toEqual(false);

    button.click();
    await flush();
    expect(requests).toHaveLength(2);
    expect(button.disabled).toEqual(true);

    // More taps do not start more requests.
    button.click();
    button.click();
    await flush();
    expect(requests).toHaveLength(2);

    // Cancel the dialog.
    requests[1].reject(notAllowed());
    await flush();
    expect(button.disabled).toEqual(false);

    button.click();
    await flush();
    expect(requests.filter((r) => r.options.mediation == null)).toHaveLength(2);
  });

  it("keeps the button disabled after a passkey is chosen", async () => {
    await render();
    const button = document.getElementById("use-passkey") as HTMLButtonElement;

    button.click();
    await flush();
    requests[1].resolve(makeCredential());
    await flush();

    expect(button.disabled).toEqual(true);
  });

  it("enables the button when the page is restored from the cache", async () => {
    await render();
    const button = document.getElementById("use-passkey") as HTMLButtonElement;

    button.click();
    await flush();
    requests[1].resolve(makeCredential());
    await flush();
    expect(button.disabled).toEqual(true);

    await showPage(false);
    expect(button.disabled).toEqual(true);

    await showPage(true);
    expect(button.disabled).toEqual(false);

    button.click();
    await flush();
    expect(button.disabled).toEqual(true);
    expect(requests.filter((r) => r.options.mediation == null)).toHaveLength(2);
  });
});

describe("PasskeyAutofillController with the WebView bridge", () => {
  for (const platform of ["ios", "android"] as const) {
    it(`does not offer autofill on ${platform}`, async () => {
      installBridge(platform, async () => ({ status: "unsupported" }));
      await render();
      await advance(60000);
      expect(requests).toHaveLength(0);

      await showPage(true);
      await advance(60000);
      expect(requests).toHaveLength(0);
    });
  }
});
