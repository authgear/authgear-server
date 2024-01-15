import { Controller } from "@hotwired/stimulus";

export interface SearchSelectOption {
  label: string;
  value: string;
}

export class CustomSelectController extends Controller {
  static targets = ["select", "itemTemplate"];
  static values = {
    options: Array,
  };

  declare readonly selectTarget: HTMLSelectElement;

  declare readonly optionsValue: SearchSelectOption[];

  // Update select options when options change
  optionsValueChanged() {
    console.log("options changed", this.optionsValue);
    this.render();
  }

  render() {
    const options = this.optionsValue;

    // Remove all options
    this.selectTarget.innerHTML = "";

    // Add new options
    options.forEach((option) => {
      const optionElement = document.createElement("option");
      optionElement.value = option.value;
      optionElement.innerText = option.label;
      this.selectTarget.appendChild(optionElement);
    });
  }
}
