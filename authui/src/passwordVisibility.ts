import { Controller } from "@hotwired/stimulus";

export class PasswordVisibilityToggleController extends Controller {
  static targets = ["input", "showButton", "hideButton"];

  declare inputTarget: HTMLInputElement;
  declare showButtonTarget: HTMLButtonElement;
  declare hideButtonTarget: HTMLButtonElement;

  connect() {
    if (this.inputTarget.type === "password") {
      this.showButtonTarget.classList.remove("hidden");
    } else {
      this.hideButtonTarget.classList.remove("hidden");
    }
  }

  show(e: Event) {
    e.preventDefault();
    e.stopImmediatePropagation();

    this.inputTarget.type = "text";
    this.showButtonTarget.classList.add("hidden");
    this.hideButtonTarget.classList.remove("hidden");
  }

  hide(e: Event) {
    e.preventDefault();
    e.stopImmediatePropagation();

    this.inputTarget.type = "password";
    this.showButtonTarget.classList.remove("hidden");
    this.hideButtonTarget.classList.add("hidden");
  }
}
