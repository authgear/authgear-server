import { Controller } from "@hotwired/stimulus";

interface Option {
  triggerLabel: string;
  searchLabel: string;
  label: string;
  value: string;
}

export class SelectInputController extends Controller {
  declare selectTarget: HTMLElement;
  declare optionsValue: Option[];

  static targets = ["select"];

  static values = {
    options: Array,
  };

  connect(): void {
    this.selectTarget.setAttribute(
      "data-custom-select-options-value",
      JSON.stringify(this.optionsValue)
    );
  }
}
