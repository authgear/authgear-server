import { Controller } from "@hotwired/stimulus";
import { visit } from "@hotwired/turbo";
import { RetryEventTarget } from "./retry";

function refreshPage() {
  let url = window.location.pathname;
  if (window.location.search !== "") {
    url += window.location.search;
  }
  if (window.location.hash !== "") {
    url += window.location.hash;
  }
  visit(url, { action: "replace" });
}

export class AuthflowWebsocketController extends Controller {
  static values = {
    url: String,
  };

  declare urlValue: string;

  ws: WebSocket | null = null;
  abortController: AbortController | null = null;
  retryEventTarget: RetryEventTarget | null = null;

  dispose = () => {
    if (this.ws != null) {
      this.ws.onclose = () => {};
      this.ws.close();
    }
    this.ws = null;
  };

  refreshIfNeeded = () => {
    refreshPage();
  };

  connectWebSocket = () => {
    const url = this.urlValue;

    this.ws = new WebSocket(url);

    this.ws.onopen = (e) => {
      console.info("authflow_websocket onopen", e);
      this.retryEventTarget?.markSuccess();
    };

    this.ws.onclose = (e) => {
      console.info("authflow_websocket onclose", e);
      // Close code 1000 means we do not need to reconnect.
      if (e.code === 1000) {
        return;
      }
      this.retryEventTarget?.scheduleRetry();
    };

    this.ws.onerror = (e) => {
      console.error("authflow_websocket onerror", e);
    };

    this.ws.onmessage = (e) => {
      console.info("authflow_websocket onmessage", e);
      const message = JSON.parse(e.data);
      switch (message.kind) {
        case "refresh":
          this.refreshIfNeeded();
      }
    };
  };

  connect() {
    this.abortController = new AbortController();
    this.retryEventTarget = new RetryEventTarget({
      abortController: this.abortController,
    });
    this.retryEventTarget.addEventListener("retry", () => {
      this.dispose();
      this.connectWebSocket();
    });

    this.connectWebSocket();
  }

  disconnect() {
    this.dispose();

    if (this.abortController != null) {
      this.abortController.abort();
    }
    this.abortController = null;

    this.retryEventTarget = null;
  }
}
