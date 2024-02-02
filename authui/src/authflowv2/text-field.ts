import { Controller } from "@hotwired/stimulus";

export class TextFieldController extends Controller {
  static values = {
    inputContainerErrorClass: { type: String },
    inputErrorClass: { type: String, default: "input--error" },
  };
  static targets = ["inputContainer", "input", "errorMessage"];

  declare inputContainerErrorClassValue: string;
  declare inputErrorClassValue: string;
  declare hasInputContainerTarget: boolean;
  declare inputContainerTarget: HTMLElement;
  declare hasErrorMessageTarget: boolean;
  declare errorMessageTarget: HTMLElement;
  declare inputTarget: HTMLInputElement;

  connect() {
    this.inputTarget.addEventListener("input", this.onInput);
  }

  disconnect() {
    this.inputTarget.removeEventListener("input", this.onInput);
  }

  onInput = () => {
    if (this.hasInputContainerTarget) {
      this.inputContainerTarget.classList.remove(
        this.inputContainerErrorClassValue
      );
    }

    if (this.inputTarget.classList.contains(this.inputErrorClassValue)) {
      this.inputTarget.classList.remove(this.inputErrorClassValue);
    }

    if (this.hasErrorMessageTarget) {
      this.errorMessageTarget.classList.add("hidden");
    }
  };
}
