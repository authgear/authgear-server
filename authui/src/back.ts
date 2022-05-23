import { Controller } from "@hotwired/stimulus";

export class BackButtonController extends Controller {
  connect() {
    this.element.classList.remove("invisible");
  }
}
