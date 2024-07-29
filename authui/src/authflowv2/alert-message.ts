import { Controller } from "@hotwired/stimulus";

export function handleAxiosError(e: unknown) {
  const err = e as any;
  if (err.response != null) {
    setErrorMessage("data-server-error-message");
  } else {
    setErrorMessage("data-network-error-message");
  }

  console.error(err);
}

export function setErrorMessage(id: string) {
  const e = new CustomEvent("alert-message:show-message", {
    detail: {
      id,
    },
  });
  document.dispatchEvent(e);
}

export class AlertMessageController extends Controller {
  static targets = ["message"];

  declare messageTarget: HTMLElement;

  // Dismiss the alert before cache the page.
  // See https://github.com/authgear/authgear-server/issues/1424
  beforeCache = () => {
    this.element.removeAttribute("data-error-message");
  };

  updateMessage = (e: Event) => {
    const errorMessage = this.element.getAttribute(
      (e as CustomEvent).detail.id
    );
    if (errorMessage != null) {
      this.element.setAttribute("data-error-message", errorMessage);
      this.displayMessage();
    }
  };

  displayMessage = () => {
    const errorMessage = this.element.getAttribute("data-error-message");
    if (errorMessage != null) {
      this.messageTarget.innerHTML = errorMessage;
    }
  };

  connect() {
    document.addEventListener("turbo:before-cache", this.beforeCache);
    document.addEventListener("alert-message:show-message", this.updateMessage);
    this.displayMessage();
  }

  disconnect() {
    document.removeEventListener("turbo:before-cache", this.beforeCache);
    document.removeEventListener(
      "alert-message:show-message",
      this.updateMessage
    );
  }
}
