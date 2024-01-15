import { Controller } from "@hotwired/stimulus";
import { countries, getCountryDataList, getEmojiFlag } from "countries-list";

export class PhoneInputController extends Controller {
  static targets = ["searchSelect"];

  declare readonly searchSelectTarget: HTMLInputElement;

  connect(): void {
    this._initPhoneCode();
  }

  _initPhoneCode() {
    const countriesData = getCountryDataList();
    const options = countriesData.map((country) => {
      return {
        label: `${getEmojiFlag(country.iso2)} +${country.phone}      ${
          country.name
        }`,
        value: country.iso2,
      };
    });
    this.searchSelectTarget.setAttribute(
      "data-custom-select-options-value",
      JSON.stringify(options)
    );
  }
}
