import Turbolinks from "turbolinks";

function refreshPage() {
  let url = window.location.pathname;
  if (window.location.search !== "") {
    url += window.location.search;
  }
  if (window.location.hash !== "") {
    url += window.location.hash;
  }
  Turbolinks.visit(url, { action: "replace" });
}

export function setupWebsocket(): () => void {
  const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
  const host = window.location.host;
  var meta: HTMLMetaElement | null = document.querySelector(
    'meta[name="x-authgear-page-loaded-at"]'
  );
  let sessionUpdatedAfter = "";
  if (meta != null) {
    sessionUpdatedAfter = meta.content || "";
  }

  let ws: WebSocket | null = null;

  function dispose() {
    if (ws != null) {
      ws.onclose = function () {};
      ws.close();
    }
    ws = null;
  }

  function refreshIfNeeded() {
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
  }

  function connect() {
    const url =
      `${scheme}//${host}/ws` +
      (sessionUpdatedAfter
        ? `?session_updated_after=${sessionUpdatedAfter}`
        : "");

    ws = new WebSocket(url);

    ws.onopen = function (e) {
      console.log("ws onopen", e);
      // after connected, we don't need to check session updated again when
      // reconnect
      // clear the checking parameter
      sessionUpdatedAfter = "";
    };

    ws.onclose = function (e) {
      console.log("ws onclose", e);
      // Close code 1000 means we do not need to reconnect.
      if (e.code === 1000) {
        return;
      }

      dispose();
      connect();
    };

    ws.onerror = function (e) {
      console.error("ws onerror", e);
    };

    ws.onmessage = function (e) {
      console.log("ws onmessage", e);
      const message = JSON.parse(e.data);
      switch (message.kind) {
        case "refresh":
          refreshIfNeeded();
      }
    };
  }

  connect();
  return dispose;
}
