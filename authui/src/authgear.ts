import Turbolinks from "turbolinks";
import { init } from "./core";
import { setupIntlTelInput } from "./intlTelInput";
import { setupPasswordPolicy } from "./password-policy";
import {
  clickLinkSubmitForm,
  autoSubmitForm,
  xhrSubmitForm,
  restoreForm,
} from "./form";
import { setupSelectEmptyValue, setupGenderSelect } from "./select";
import { formatDateRelative, formatInputDate } from "./date";
import { setupAccountDeletion } from "./accountdeletion";
// FIXME(css): Build CSS files one by one with another tool
// webpack bundles all CSS files into one bundle.

init();

window.api.onLoad(() => {
  document.body.classList.add("js");
});

window.api.onLoad(setupPasswordPolicy);

window.api.onLoad(setupIntlTelInput);

window.api.onLoad(setupSelectEmptyValue);
window.api.onLoad(setupGenderSelect);

window.api.onLoad(formatDateRelative);
window.api.onLoad(formatInputDate);

window.api.onLoad(setupAccountDeletion);

function copyToClipboard(str: string): void {
  const el = document.createElement("textarea");
  el.value = str;
  // Set non-editable to avoid focus and move outside of view
  el.setAttribute("readonly", "");
  el.setAttribute("style", "position: absolute; left: -9999px");
  document.body.appendChild(el);
  // Select text inside element
  el.select();
  el.setSelectionRange(0, el.value.length); // for mobile device
  document.execCommand("copy");
  // Remove temporary element
  document.body.removeChild(el);
}

// Disable double tap zoom
// There are rumours on the web claiming that touch-action: manipulation can do that.
// However, I tried
// * {
//   touch-action: manipulation;
// }
// and
// * {
//   touch-action: pan-y;
// }
// None of them work.
window.api.onLoad(() => {
  function listener(e: Event) {
    e.preventDefault();
    e.stopPropagation();
  }
  document.addEventListener("dblclick", listener);
  return () => {
    document.removeEventListener("dblclick", listener);
  };
});

// Copy button
window.api.onLoad(() => {
  function copy(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    const button = e.currentTarget as HTMLElement;
    const targetSelector = button.getAttribute("data-copy-button-target");
    if (targetSelector == null) {
      return;
    }

    const copyLabel = button.getAttribute("data-copy-button-copy-label");
    const copiedLabel = button.getAttribute("data-copy-button-copied-label");

    const target = document.querySelector(targetSelector);
    if (target == null) {
      return;
    }

    const textContent = target.textContent;
    if (textContent == null) {
      return;
    }

    copyToClipboard(textContent);

    // Show feedback
    const currentHandle = button.getAttribute(
      "data-copy-button-timeout-handle"
    );
    // Clear scheduled timeout if the timeout function has NOT been executed yet.
    if (currentHandle != null) {
      window.clearTimeout(Number(currentHandle));
      button.removeAttribute("data-copy-button-timeout-handle");
    }
    // Changing label as feedback is optional
    if (copyLabel != null && copiedLabel != null) {
      button.textContent = copiedLabel;
    }
    button.classList.add("outline");
    const newHandle = window.setTimeout(() => {
      // Changing label as feedback is optional
      if (copyLabel != null && copiedLabel != null) {
        button.textContent = copyLabel;
      }
      button.classList.remove("outline");
      button.removeAttribute("data-copy-button-timeout-handle");
    }, 1000);
    button.setAttribute("data-copy-button-timeout-handle", String(newHandle));
  }

  const elems = document.querySelectorAll("[data-copy-button-target]");
  const buttons: HTMLElement[] = [];
  for (let i = 0; i < elems.length; i++) {
    const elem = elems[i];
    if (elem instanceof HTMLElement) {
      buttons.push(elem);
    }
  }
  for (const button of buttons) {
    button.addEventListener("click", copy);
  }
  return () => {
    for (const button of buttons) {
      button.removeEventListener("click", copy);
    }
  };
});

// Handle message bar close button
window.api.onLoad(() => {
  const disposers: Array<() => void> = [];
  const closeButtons = document.querySelectorAll("[data-close-button-target]");

  for (let i = 0; i < closeButtons.length; i++) {
    const closeButton = closeButtons[i];

    const targetID = closeButton.getAttribute("data-close-button-target");
    if (targetID == null) {
      continue;
    }

    const target = document.getElementById(targetID);
    if (target == null) {
      continue;
    }

    const onCloseButtonClick = (e: Event) => {
      e.preventDefault();
      e.stopPropagation();
      target.classList.add("hidden");
    };

    // Close the message bar before cache the page.
    // So that the cached page does not have the message bar shown.
    // See https://github.com/authgear/authgear-server/issues/1424
    const beforeCache = () => {
      if (closeButton instanceof HTMLElement) {
        closeButton.click();
      }
    };

    closeButton.addEventListener("click", onCloseButtonClick);
    document.addEventListener("turbolinks:before-cache", beforeCache);
    disposers.push(() => {
      closeButton.removeEventListener("click", onCloseButtonClick);
      document.removeEventListener("turbolinks:before-cache", beforeCache);
    });
  }

  return () => {
    for (const disposer of disposers) {
      disposer();
    }
  };
});

window.api.onLoad(xhrSubmitForm);
window.api.onLoad(restoreForm);

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

// Handle password visibility toggle.
window.api.onLoad(() => {
  const passwordInputs = document.querySelectorAll(
    "[data-show-password-button]"
  );

  const disposers: Array<() => void> = [];

  for (let i = 0; i < passwordInputs.length; i++) {
    const passwordInput = passwordInputs[i] as HTMLInputElement;

    const showPasswordButtonID = passwordInput.getAttribute(
      "data-show-password-button"
    );
    const hidePasswordButtonID = passwordInput.getAttribute(
      "data-hide-password-button"
    );
    if (showPasswordButtonID == null || hidePasswordButtonID == null) {
      continue;
    }

    const showPasswordButton = document.getElementById(showPasswordButtonID);
    const hidePasswordButton = document.getElementById(hidePasswordButtonID);
    if (showPasswordButton == null || hidePasswordButton == null) {
      continue;
    }

    const togglePasswordVisibility = (e: Event) => {
      e.preventDefault();
      e.stopPropagation();

      if (hidePasswordButton.classList.contains("hidden")) {
        passwordInput.type = "text";
        showPasswordButton.classList.add("hidden");
        hidePasswordButton.classList.remove("hidden");
      } else {
        passwordInput.type = "password";
        showPasswordButton.classList.remove("hidden");
        hidePasswordButton.classList.add("hidden");
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

window.api.onLoad(autoSubmitForm);
window.api.onLoad(clickLinkSubmitForm);

// Handle click link switch label and href
window.api.onLoad(() => {
  const targets = document.querySelectorAll("[data-switch-to-on-click]");

  function listener(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    if (!(e.currentTarget instanceof HTMLElement)) {
      return;
    }
    const selector = e.currentTarget.getAttribute("data-switch-to-on-click");
    if (selector == null) {
      return;
    }

    const selectedElement = document.querySelector(selector);
    if (selectedElement == null) {
      return;
    }

    e.currentTarget.classList.add("hidden");
    selectedElement.classList.remove("hidden");
  }

  for (let i = 0; i < targets.length; i++) {
    const target = targets[i];
    target.addEventListener("click", listener);
  }

  return () => {
    for (let i = 0; i < targets.length; i++) {
      const target = targets[i];
      target.removeEventListener("click", listener);
    }
  };
});

// Handle confirmation modal
// usage: adding follow data attribute in the button element
// - data-modal="confirmation"
// - data-modal-title="{TITLE_TEXT}"
// - data-modal-body="{BODY_TEXT}"
// - data-modal-action-label="{ACTION_LABEL_TEXT}"
// - data-modal-cancel-label="{CANCEL_LABEL_TEXT}"
window.api.onLoad(() => {
  const modal = document.querySelector('[data-modal-ele="true"]');
  if (!(modal instanceof HTMLElement)) {
    // modal template not found
    return;
  }

  const modalTitleEle = modal.querySelector(
    '[data-modal-title-ele="true"]'
  ) as HTMLElement;
  const modalBodyEle = modal.querySelector(
    '[data-modal-body-ele="true"]'
  ) as HTMLElement;
  const modalActionBtnEle = modal.querySelector(
    '[data-modal-action-btn-ele="true"]'
  ) as HTMLElement;
  const modalCancelBtnEle = modal.querySelector(
    '[data-modal-cancel-btn-ele="true"]'
  ) as HTMLElement;
  const modalOverlayEle = modal.querySelector(
    '[data-modal-overlay-ele="true"]'
  ) as HTMLElement;

  const buttons = document.querySelectorAll('[data-modal="confirmation"]');
  const disposers: Array<() => void> = [];
  var confirmed = false;

  for (let i = 0; i < buttons.length; i++) {
    const button = buttons[i] as HTMLElement;

    const closeModal = () => {
      confirmed = false;
      disposeModal();
      modal.classList.add("closed");
    };

    const onClickModalAction = (e: Event) => {
      confirmed = true;
      button.click();
    };

    const onClickModalCancel = (e: Event) => {
      closeModal();
    };

    const disposeModal = () => {
      modalActionBtnEle.removeEventListener("click", onClickModalAction);
      modalCancelBtnEle.removeEventListener("click", onClickModalCancel);
      modalOverlayEle.removeEventListener("click", onClickModalCancel);
    };

    const confirmFormSubmit = (e: Event) => {
      if (confirmed) {
        // close the modal and perform the default behaviour
        closeModal();
        return;
      }
      e.preventDefault();
      modalTitleEle.innerText = button.dataset["modalTitle"] || "";
      modalBodyEle.innerText = button.dataset["modalBody"] || "";
      modalActionBtnEle.innerText = button.dataset["modalActionLabel"] || "";
      modalCancelBtnEle.innerText = button.dataset["modalCancelLabel"] || "";

      modalActionBtnEle.addEventListener("click", onClickModalAction);
      modalCancelBtnEle.addEventListener("click", onClickModalCancel);
      modalOverlayEle.addEventListener("click", onClickModalCancel);

      modal.classList.remove("closed");
    };

    button.addEventListener("click", confirmFormSubmit);
    disposers.push(() => {
      button.removeEventListener("click", confirmFormSubmit);
      disposeModal();
    });
  }

  return () => {
    for (const disposer of disposers) {
      disposer();
    }
  };
});

// Websocket runtime
window.api.onLoad(() => {
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
});
