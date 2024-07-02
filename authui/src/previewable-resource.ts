import { Controller } from "@hotwired/stimulus";

export class PreviewableResourceController extends Controller {
  static values = {
    key: String,
    changableAttribute: String,
    original: String,
  };

  declare keyValue: string;
  declare changableAttributeValue: string;
  declare originalValue: string;

  setMessage(message: string): void {
    this.element.innerHTML = message;
  }

  setValue(value: string | null): void {
    const valueToSet = value != null ? value : this.originalValue;
    if (this.changableAttributeValue === "innerHTML") {
      this.element.innerHTML = valueToSet;
    } else {
      this.element.setAttribute(this.changableAttributeValue, valueToSet);
    }
  }
}
