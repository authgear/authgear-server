import { Controller } from "@hotwired/stimulus";

export class TextFieldController extends Controller {
  static values = {
    inputErrorClass: { type: String, default: "input--error" },
  };
  static targets = ["input", "errorMessage"];

  declare inputErrorClassValue: string;
  declare errorMessageTarget: HTMLElement | null;
  declare inputTarget: HTMLInputElement;

  connect() {
    this.inputTarget.addEventListener("input", this.onInput);
  }

  disconnect() {
    this.inputTarget.removeEventListener("input", this.onInput);
  }

  onInput = () => {
    if (this.inputTarget.classList.contains(this.inputErrorClassValue)) {
      this.inputTarget.classList.remove(this.inputErrorClassValue);
    }

    if (this.errorMessageTarget != null) {
      this.errorMessageTarget.style.display = "none";
    }
  };
}
