import { Controller } from "@hotwired/stimulus";
import { countries, getCountryDataList, getEmojiFlag } from "countries-list";

export class PhoneInputController extends Controller {
  static targets = ["customSelect"];

  declare readonly customSelectTarget: HTMLInputElement;

  connect(): void {
    this._initPhoneCode();
  }

  _initPhoneCode() {
    const countriesData = getCountryDataList().sort((a, b) => {
      return a.name.localeCompare(b.name);
    });
    const options = countriesData.map((country) => {
      return {
        triggerLabel: `${getEmojiFlag(country.iso2)} +${country.phone}`,
        label: `${getEmojiFlag(country.iso2)} +${country.phone}       ${
          country.name
        }`,
        value: country.iso2,
      };
    });
    this.customSelectTarget.setAttribute(
      "data-custom-select-options-value",
      JSON.stringify(options)
    );
  }
}
