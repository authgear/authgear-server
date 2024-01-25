import { Controller } from "@hotwired/stimulus";
import { disableAllButtons } from "../loading";

interface LoadingHandle {
  onError: () => void;
  onFinally: () => void;
}

export class LoadingController extends Controller {
  revert: (() => void) | null = null;

  // Revert disabled buttons before Turbo caches the page
  // To avoid flickering in the UI
  beforeCache = () => {
    this.revert?.();
    this.revert = null;
  };

  startLoading(e: HTMLElement | null): LoadingHandle {
    this.revert = disableAllButtons();
    e?.setAttribute("data-loading", "true");

    return {
      onError: () => {
        this.revert?.();
        this.revert = null;
        e?.removeAttribute("data-loading");
      },
      onFinally: () => {
        e?.removeAttribute("data-loading");
      },
    };
  }

  connect() {
    document.addEventListener("turbo:before-cache", this.beforeCache);
  }

  disconnect() {
    document.removeEventListener("turbo:before-cache", this.beforeCache);
  }
}
