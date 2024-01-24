import { Controller } from "@hotwired/stimulus";
import {
  disableAllButtons,
  hideProgressBar,
  showProgressBar,
} from "../loading";

interface LoadingEvent extends Event {
  detail?: {
    button?: HTMLButtonElement;
    error?: unknown;
  };
}

export class LoadingController extends Controller {
  revertDisabledButtons: { (): void } | null = null;
  button: HTMLButtonElement | null = null;

  // Revert disabled buttons before Turbo caches the page
  // To avoid flickering in the UI
  beforeCache = () => {
    if (this.revertDisabledButtons) {
      this.revertDisabledButtons();
    }
  };

  onLoading(e: LoadingEvent) {
    this.revertDisabledButtons = disableAllButtons();
    showProgressBar();
    if (e.detail?.button?.getAttribute("data-loading-state") != null) {
      e.detail.button.classList.add("primary-btn--loading");
      this.button = e.detail.button;
    }
  }
  onLoadingError(e: LoadingEvent) {
    // revert is only called for error branch because
    // The success branch also loads a new page.
    // Keeping the buttons in disabled state reduce flickering in the UI.
    if (this.revertDisabledButtons) {
      this.revertDisabledButtons();
      this.revertDisabledButtons = null;
    }
    this.button?.classList.remove("primary-btn--loading");
  }
  onLoadingEnd(e: LoadingEvent) {
    hideProgressBar();
    this.button?.classList.remove("primary-btn--loading");
  }

  connect() {
    this.element.addEventListener("onLoading", this.onLoading);
    this.element.addEventListener("onLoadingError", this.onLoadingError);
    this.element.addEventListener("onLoadingEnd", this.onLoadingEnd);
    document.addEventListener("turbo:before-cache", this.beforeCache);
  }

  disconnect() {
    this.element.removeEventListener("onLoading", this.onLoading);
    this.element.removeEventListener("onLoadingError", this.onLoadingError);
    this.element.removeEventListener("onLoadingEnd", this.onLoadingEnd);
    document.removeEventListener("turbo:before-cache", this.beforeCache);
  }
}
