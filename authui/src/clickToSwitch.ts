import { Controller } from "@hotwired/stimulus";

export class ClickToSwitchController extends Controller {
  static targets = ["clickToShow", "clickToHide"];

  declare clickToShowTarget: HTMLElement;
  declare clickToHideTarget: HTMLElement;

  click() {
    // Do not call prevent default intentionally.
    this.clickToHideTarget.classList.add("hidden");
    this.clickToShowTarget.classList.remove("hidden");
  }
}
