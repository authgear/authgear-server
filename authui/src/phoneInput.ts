import { Controller } from "@hotwired/stimulus";
import {
  ICountryData,
  countries,
  getCountryDataList,
  getEmojiFlag,
} from "countries-list";

export class PhoneInputController extends Controller {
  static targets = ["countrySelect", "input"];

  declare readonly countrySelectTarget: HTMLSelectElement;
  declare readonly inputTarget: HTMLInputElement;

  countryCode?: string;
  phoneNumber?: string;

  _countriesData: ICountryData[] = [];

  connect(): void {
    this._initPhoneCode();
  }

  updateValue(): void {
    const country = this._countriesData.find(
      (country) => country.iso2 === this.countryCode
    );
    const phoneCode = country?.phone[0];
    const value =
      phoneCode && this.phoneNumber
        ? `+${country?.phone[0]}${this.phoneNumber}`
        : "";
    this.inputTarget.value = value;
  }

  handleNumberInput(event: Event): void {
    const target = event.target as HTMLInputElement;
    target.value = target.value.replace(/\D/g, "");
    const value = target.value;
    this.phoneNumber = value;
    this.updateValue();
  }

  handleCountryInput(event: Event): void {
    const target = event.target as HTMLInputElement;
    const value = target.value;
    this.countryCode = value;
    this.updateValue();
  }

  _initPhoneCode() {
    this._countriesData = getCountryDataList().sort((a, b) => {
      return a.name.localeCompare(b.name);
    });
    const options = this._countriesData.map((country) => {
      return {
        triggerLabel: `${getEmojiFlag(country.iso2)} +${country.phone}`,
        prefix: `${getEmojiFlag(country.iso2)} +${country.phone}`,
        label: country.name,
        value: country.iso2,
      };
    });
    this.countryCode = options[0].value;
    this.countrySelectTarget.setAttribute(
      "data-custom-select-options-value",
      JSON.stringify(options)
    );
    this.updateValue();
  }
}
