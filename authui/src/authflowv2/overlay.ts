import { Controller } from "@hotwired/stimulus";

export class OverlayController extends Controller {
  static values = { defaultOpen: Boolean };

  declare readonly defaultOpenValue: boolean;

  connect(): void {
    if (this.defaultOpenValue) {
      this.open();
    }
  }

  open() {
    this.element.classList.add("open");
  }

  close() {
    this.element.classList.remove("open");
  }
}
