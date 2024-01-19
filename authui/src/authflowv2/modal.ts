import { Controller } from "@hotwired/stimulus";

export class ModalController extends Controller {
  close() {
    (this.element as HTMLDialogElement).removeAttribute("open");
    (this.element as HTMLDialogElement).close();
  }
}
