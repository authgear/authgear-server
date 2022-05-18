import { Controller } from "@hotwired/stimulus";

export class AccountDelectionController extends Controller {
  static targets = ["input", "button"];

  declare inputTarget: HTMLInputElement;
  declare buttonTarget: HTMLButtonElement;

  delete() {
    const input = this.inputTarget;
    const button = this.buttonTarget;

    const inputText = input.value;
    button.disabled = inputText !== "DELETE";
  }
}
