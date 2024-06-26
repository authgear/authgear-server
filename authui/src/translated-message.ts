import { Controller } from "@hotwired/stimulus";

export class TranslatedMessageController extends Controller {
  static values = {
    key: String,
  };

  declare keyValue: string;

  setMessage(message: string): void {
    this.element.innerHTML = message;
  }
}
