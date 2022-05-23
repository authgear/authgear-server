import { Controller } from "@hotwired/stimulus";

export class BackButtonController extends Controller {
  connect() {
    const button = this.element as HTMLButtonElement;

    button.style.visibility = "visible";
  }
}
