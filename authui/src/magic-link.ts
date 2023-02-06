import { Controller } from "@hotwired/stimulus";

export class MagicLinkAutoVerifyController extends Controller {
  static targets = ["input", "submit"];

  declare inputTarget: HTMLInputElement;
  declare submitTarget: HTMLButtonElement;

  onVerify(event: CustomEvent) {
    this.inputTarget.value = event.detail;
    this.submitTarget.click();
  }
}
