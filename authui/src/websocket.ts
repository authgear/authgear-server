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

export class WebSocketController extends Controller {
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
    const ele = document.querySelector('[data-is-refresh-link="true"]');
    if (ele) {
      // if there is refresh link in the page, don't refresh automatically
      return;
    }
    const btn = document.querySelector('[data-submit-when-refresh="true"]');
    if (btn instanceof HTMLElement) {
      btn.click();
      return;
    }

    refreshPage();
  };

  connectWebSocket = (isReconnect: boolean) => {
    const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    var meta: HTMLMetaElement | null = document.querySelector(
      'meta[name="x-authgear-page-loaded-at"]'
    );
    let sessionUpdatedAfter = "";
    if (meta != null) {
      sessionUpdatedAfter = meta.content || "";
    }

    // We only pass session_updated_after in case of reconnection.
    // If we also pass session_updated_after in first connection,
    // we will receive a refresh message which will refresh the page immediately.
    // This will cause the page to load twice.
    const url =
      `${scheme}//${host}/ws` +
      (isReconnect && sessionUpdatedAfter != ""
        ? `?session_updated_after=${sessionUpdatedAfter}`
        : "");

    this.ws = new WebSocket(url);

    this.ws.onopen = (e) => {
      console.log("ws onopen", e);
      this.retryEventTarget?.markSuccess();
    };

    this.ws.onclose = (e) => {
      console.log("ws onclose", e);
      // Close code 1000 means we do not need to reconnect.
      if (e.code === 1000) {
        return;
      }
      this.retryEventTarget?.scheduleRetry();
    };

    this.ws.onerror = (e) => {
      console.error("ws onerror", e);
    };

    this.ws.onmessage = (e) => {
      console.log("ws onmessage", e);
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
      this.connectWebSocket(true);
    });

    this.connectWebSocket(false);
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
