import { Controller } from "@hotwired/stimulus";

export class DialogController extends Controller {
  close() {
    (this.element as HTMLDialogElement).close();
  }
}
