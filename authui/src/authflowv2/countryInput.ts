import { Controller } from "@hotwired/stimulus";

interface CountryOption {
  code: string;
  displayLabel: string;
}

export class CountryInputController extends Controller {
  declare countrySelectTarget: HTMLElement;
  declare countryOptionsValue: CountryOption[];

  static targets = ["countrySelect"];

  static values = {
    countryOptions: Array,
  };

  connect(): void {
    const selectOptions = this.countryOptionsValue.map((option) => ({
      triggerLabel: option.displayLabel,
      searchLabel: `${option.code} ${option.displayLabel}`,
      label: option.displayLabel,
      value: option.code,
    }));
    this.countrySelectTarget.setAttribute(
      "data-custom-select-options-value",
      JSON.stringify(selectOptions)
    );
  }
}
