import Turbolinks from "turbolinks";
import { init } from "./core";
import "./authgear.css";

init();

window.api.onLoad(() => {
  document.body.classList.add("js");
});

// Handle history tracking.

let rootArea: "settings" | null;
let historyRootSet = false;
window.api.onLoad(() => {
  if (!historyRootSet) {
    const meta = document.createElement("meta");
    meta.name = "authgear-history-root"
    document.head.appendChild(meta);
    historyRootSet = true;
  }
  if (window.location.pathname.startsWith("/settings")) {
    rootArea = "settings";
  } else {
    rootArea = null;
  }
});

// Handle form submission

function setNetworkError() {
  const field = document.querySelector(".errors");
  if (field) {
    field.textContent = field.getAttribute("data-network-error");
  }
}

function setServerError() {
  const field = document.querySelector(".errors");
  if (field) {
    field.textContent = field.getAttribute("data-server-error");
  }
}

window.api.onLoad(() => {
  let isSubmitting = false;
  function submitForm(e: Event) {
    e.preventDefault();
    e.stopPropagation();
    if (isSubmitting) {
      return;
    }
    isSubmitting = true;

    const form = e.currentTarget as HTMLFormElement;
    const formData = new FormData(form);

    const params = new URLSearchParams();
    formData.forEach((value, name) => {
      params.set(name, value as string);
    });
    // FormData does not include any submit button's data:
    // include them manually, since we have at most one submit button per form.
    const submitButtons = form.querySelectorAll('button[type="submit"]');
    for (let i = 0; i < submitButtons.length; i++) {
      const button = submitButtons[i] as HTMLButtonElement;
      params.set(button.name, button.value);
    }
    if (form.id) {
      const el = document.querySelector(
        `button[type="submit"][form="${form.id}"]`
      );
      if (el) {
        const button = el as HTMLButtonElement;
        params.set(button.name, button.value);
      }
    }

    fetch(form.action, {
      method: form.method,
      headers: {
        "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
        "X-Authgear-XHR": "true"
      },
      credentials: "same-origin",
      body: params
    })
      .then(resp => {
        if (resp.status < 200 || resp.status >= 300) {
          isSubmitting = false;
          setServerError();
          return;
        }
        return resp.json().then(({ redirect_uri, action }) => {
          isSubmitting = false;

          Turbolinks.clearCache();
          switch (action) {
            case "redirect":
              // Perform full redirect.
              window.location = redirect_uri;
              break;

            case "replace":
            case "advance":
              Turbolinks.visit(redirect_uri, { action });
              break;
          }
        });
      })
      .catch(() => {
        isSubmitting = false;
        setNetworkError();
      });
  }

  const elems = document.querySelectorAll("form");
  const forms: HTMLFormElement[] = [];
  for (let i = 0; i < elems.length; i++) {
    if (elems[i].querySelector('[data-form-xhr="false"]')) {
      continue;
    }
    forms.push(elems[i] as HTMLFormElement);
  }
  for (const form of forms) {
    form.addEventListener("submit", submitForm);
  }

  return () => {
    for (const form of forms) {
      form.removeEventListener("submit", submitForm);
    }
  };
});

// Handle back button.

function back(e: Event) {
  e.preventDefault();
  e.stopPropagation();
  const rootMeta = document.querySelector('meta[name="authgear-history-root"]');
  if (rootMeta) {
    if (rootArea === "settings" && window.location.pathname !== "/settings") {
      // Replace the history root with settings top-page.
      // FIXME: forward history is not cleared.
      rootMeta.remove();
      Turbolinks.visit("/settings", { action: "replace" });
      historyRootSet = false;
      return;
    }
  }
  window.history.back();
}

function refreshPage() {
  let url = window.location.pathname;
  if (window.location.search !== "") {
    url += "?" + window.location.search;
  }
  if (window.location.hash !== "") {
    url += "#" + window.location.hash;
  }
  Turbolinks.visit(url, { action: "replace" });
}

window.api.onLoad(() => {
  const elems = document.querySelectorAll(".back-btn");
  for (let i = 0; i < elems.length; i++) {
    elems[i].addEventListener("click", back);
  }

  return () => {
    for (let i = 0; i < elems.length; i++) {
      elems[i].removeEventListener("click", back);
    }
  };
});

// Handle password visibility toggle.

window.api.onLoad(() => {
  const wrappers = document.querySelectorAll(".password-input-wrapper");
  const disposers: Array<() => void> = [];
  for (let i = 0; i < wrappers.length; i++) {
    const wrapper = wrappers[i];
    const input = wrapper.querySelector(".input") as HTMLInputElement;
    const showPasswordButton = wrapper.querySelector(".show-password-button");
    const hidePasswordButton = wrapper.querySelector(".hide-password-button");
    if (!input || !showPasswordButton || !hidePasswordButton) {
      return;
    }

    if (wrapper.classList.contains("show-password")) {
      input.type = "text";
    } else {
      input.type = "password";
    }

    const togglePasswordVisibility = (e: Event) => {
      e.preventDefault();
      e.stopPropagation();
      wrapper.classList.toggle("show-password");
      if (wrapper.classList.contains("show-password")) {
        input.type = "text";
      } else {
        input.type = "password";
      }
    };

    showPasswordButton.addEventListener("click", togglePasswordVisibility);
    hidePasswordButton.addEventListener("click", togglePasswordVisibility);
    disposers.push(() => {
      showPasswordButton.removeEventListener("click", togglePasswordVisibility);
      hidePasswordButton.removeEventListener("click", togglePasswordVisibility);
    });
  }

  return () => {
    for (const disposer of disposers) {
      disposer();
    }
  };
});

// Handle resend button.

window.api.onLoad(() => {
  const el = document.querySelector("#resend-button") as HTMLButtonElement;
  if (el == null) {
    return;
  }

  const scheduledAt = new Date();
  const cooldown = Number(el.getAttribute("data-cooldown")) * 1000;
  const label = el.getAttribute("data-label");
  const labelUnit = el.getAttribute("data-label-unit")!;
  let animHandle: number | null = null;

  function tick() {
    const now = new Date();
    const timeElapsed = now.getTime() - scheduledAt.getTime();

    let displaySeconds = 0;
    if (timeElapsed <= cooldown) {
      displaySeconds = Math.round((cooldown - timeElapsed) / 1000);
    }

    if (displaySeconds === 0) {
      el.disabled = false;
      el.textContent = label;
      animHandle = null;
    } else {
      el.disabled = true;
      el.textContent = labelUnit.replace("%d", String(displaySeconds));
      animHandle = requestAnimationFrame(tick);
    }
  }

  animHandle = requestAnimationFrame(tick);

  return () => {
    if (animHandle != null) {
      cancelAnimationFrame(animHandle);
    }
  };
});

window.api.onLoad(() => {
  const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
  const host = window.location.host;
  const url = `${scheme}//${host}/ws`;

  let ws: WebSocket | null = null;

  function dispose() {
    ws?.close();
    ws = null;
  }

  function connect() {
    ws = new WebSocket(url);

    ws.onopen = function(e) {
      console.log("ws onopen", e);
    };

    ws.onclose = function(e) {
      console.log("ws onclose", e);
      // Close code 1000 means we do not need to reconnect.
      if (e.code === 1000) {
        return;
      }

      dispose();
      connect();
    };

    ws.onerror = function(e) {
      console.error("ws onerror", e);
    };

    ws.onmessage = function(e) {
      console.log("ws onmessage", e);
      const message = JSON.parse(e.data);
      switch (message.kind) {
      case "refresh":
        refreshPage();
      }
    }
  }

  // Avoid websocket connection unless it is necessary.
  if (window.location.pathname.startsWith("/verify_identity")) {
    connect();
  }

  if (window.location.pathname.startsWith("/sso/wechat/auth")) {
    connect();
  }

  return dispose;
});
