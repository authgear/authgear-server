import { Controller } from "@hotwired/stimulus";

interface LanguageOption {
  code: string;
  displayLabel: string;
}

export class LocaleInputController extends Controller {
  declare languagesValue: LanguageOption[];

  static targets = ["localeSelect"];
  static values = {
    languages: Array,
  };

  declare readonly localeSelectTarget: HTMLElement;

  connect() {
    const selectOptions = this.languagesValue.map((language) => ({
      triggerLabel: language.displayLabel,
      searchLabel: `${language.code} ${language.displayLabel}`,
      label: language.code === "" ? "-" : language.displayLabel,
      value: language.code,
    }));
    this.localeSelectTarget.setAttribute(
      "data-custom-select-options-value",
      JSON.stringify(selectOptions)
    );
  }
}
