import { Controller } from "@hotwired/stimulus";

interface Timezone {
  name: string;
  formattedOffset: string;
  displayLabel: string;
}

export class TimezoneInput extends Controller {
  declare timezoneSelectTarget: HTMLElement;
  declare timezonesValue: Timezone[];

  static targets = ["timezoneSelect"];

  static values = {
    timezones: Array,
  };

  connect(): void {
    const selectOptions = this.timezonesValue.map((timezone) => ({
      triggerLabel: timezone.displayLabel,
      searchLabel: `${timezone.name} ${timezone.displayLabel}`,
      label: timezone.displayLabel,
      value: timezone.name,
    }));
    this.timezoneSelectTarget.setAttribute(
      "data-custom-select-options-value",
      JSON.stringify(selectOptions)
    );
  }
}
