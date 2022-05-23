import { Controller } from "@hotwired/stimulus";

export class MessageBarController extends Controller {
  static targets = ["bar"];

  declare barTarget: HTMLElement;

  // Close the message bar before cache the page.
  // So that the cached page does not have the message bar shown.
  // See https://github.com/authgear/authgear-server/issues/1424
  beforeCache = () => {
    this.hide();
  };

  hide = () => {
    const barTarget = this.barTarget;
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
}
