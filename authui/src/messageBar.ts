import { Controller } from "@hotwired/stimulus";

export class MessageBarController extends Controller {
  static targets = ["button", "bar"];

  declare buttonTarget: HTMLButtonElement;
  declare barTarget: HTMLElement;

  // Close the message bar before cache the page.
  // So that the cached page does not have the message bar shown.
  // See https://github.com/authgear/authgear-server/issues/1424
  beforeCache = () => {
    const button = this.buttonTarget;
    button.click();
  };

  connect() {
    document.addEventListener("turbolinks:before-cache", this.beforeCache);
  }

  close(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    const barTarget = this.barTarget;

    barTarget.classList.add("hidden");
  }

  disconnect() {
    document.removeEventListener("turbolinks:before-cache", this.beforeCache);
  }
}
