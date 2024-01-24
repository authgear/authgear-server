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
      e.detail.button.innerHTML =
        '<span class="primary-btn__loading-icon material-icons animate-spin">progress_activity</span>';
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
  }
  onLoadingEnd(e: LoadingEvent) {
    hideProgressBar();
  }

  connect() {
    this.element.addEventListener("onLoading", this.onLoading);
    this.element.addEventListener("onLoadingError", this.onLoadingError);
    this.element.addEventListener("onLoadingEnd", this.onLoadingEnd);
  }

  disconnect() {
    this.element.removeEventListener("onLoading", this.onLoading);
    this.element.removeEventListener("onLoadingError", this.onLoadingError);
    this.element.removeEventListener("onLoadingEnd", this.onLoadingEnd);
  }
}
