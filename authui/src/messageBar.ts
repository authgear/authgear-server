import { AxiosResponse } from "axios";
import { Controller } from "@hotwired/stimulus";
import { session } from "@hotwired/turbo";

export function handleAxiosError(e: unknown) {
  const err = e as any;
  if (err.code === "ERR_NETWORK") {
    showErrorMessage("error-message-network");
  } else if (err.response != null) {
    const response: AxiosResponse = err.response;

    if (typeof response.data === "string") {
      session.renderStreamMessage(response.data);
      return;
    }

    showErrorMessage("error-message-server");
  } else if (err.request != null) {
    showErrorMessage("error-message-network");
  } else {
    // programming error
  }

  console.error(err);
}

export function showErrorMessage(id: string) {
  setErrorMessage(id, false);
}

export function hideErrorMessage(id: string) {
  setErrorMessage(id, true);
}

function setErrorMessage(id: string, hidden: boolean) {
  if (hidden) {
    const e = new CustomEvent("messagebar:hide-message", {
      detail: {
        id,
      },
    });
    document.dispatchEvent(e);
  } else {
    const e = new CustomEvent("messagebar:show-message", {
      detail: {
        id,
      },
    });
    document.dispatchEvent(e);
  }
}

export class MessageBarController extends Controller {
  // Close the message bar before cache the page.
  // So that the cached page does not have the message bar shown.
  // See https://github.com/authgear/authgear-server/issues/1424
  beforeCache = () => {
    this.hide();
  };

  open = () => {
    const barTarget = this.element as HTMLElement;
    barTarget.classList.remove("hidden");
  };

  hide = () => {
    const barTarget = this.element as HTMLElement;
    barTarget.classList.add("hidden");
  };

  connect() {
    document.addEventListener("turbo:before-cache", this.beforeCache);
  }

  close(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    this.hide();
  }

  disconnect() {
    document.removeEventListener("turbo:before-cache", this.beforeCache);
  }

  showMessage(e: CustomEvent) {
    this.open();
    const id: string = e.detail.id;
    document.getElementById(id)?.classList?.remove("hidden");
  }

  hideMessage(e: CustomEvent) {
    this.hide();
    const id: string = e.detail.id;
    document.getElementById(id)?.classList?.add("hidden");
  }
}
