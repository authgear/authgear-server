import { Controller } from "@hotwired/stimulus";
import { CountryCode, getCountryCallingCode } from "libphonenumber-js";
import defaultTerritories from "cldr-localenames-full/main/en/territories.json";
import territoriesMap from "cldr-localenames-full/main/*/territories.json";
import { getEmojiFlag } from "./getEmojiFlag";

interface PhoneInputCountry {
  flagEmoji: string;
  localizedName: string;
  name: string;
  iso2: string;
  phone: string;
}

export class PhoneInputController extends Controller {
  static targets = ["countrySelect", "input"];

  declare readonly countrySelectTarget: HTMLSelectElement;
  declare readonly inputTarget: HTMLInputElement;

  countryCode?: string;
  phoneNumber?: string;

  _countries: PhoneInputCountry[] = [];

  connect(): void {
    this._initPhoneCode();
  }

  updateValue(): void {
    const country = this._countries.find(
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

  async _initPhoneCode() {
    const onlyCountries: CountryCode[] =
      JSON.parse(
        document
          .querySelector("meta[name=x-phone-input-only-countries]")
          ?.getAttribute("content") ?? "null"
      ) ?? [];

    const preferredCountries: CountryCode[] =
      JSON.parse(
        document
          .querySelector("meta[name=x-phone-input-preferred-countries]")
          ?.getAttribute("content") ?? "null"
      ) ?? [];
    let initialCountry: CountryCode | null =
      (document
        .querySelector("meta[name=x-geoip-country-code]")
        ?.getAttribute("content") as CountryCode) ?? null;

    const lang = document.documentElement.lang || "en";
    const countryCodesMap: Map<CountryCode, null> = new Map();
    for (const preferredCountry of preferredCountries) {
      if (countryCodesMap.has(preferredCountry)) {
        continue;
      }
      countryCodesMap.set(preferredCountry, null);
    }
    for (const onlyCountry of onlyCountries) {
      if (countryCodesMap.has(onlyCountry)) {
        continue;
      }
      countryCodesMap.set(onlyCountry, null);
    }
    const countryCodes: CountryCode[] = Array.from(countryCodesMap.keys());

    const localizedTerritories = territoriesMap[lang];
    const territories =
      localizedTerritories?.main[lang as keyof typeof localizedTerritories.main]
        ?.localeDisplayNames.territories ||
      defaultTerritories.main.en.localeDisplayNames.territories;

    this._countries = countryCodes.map((countryCode) => {
      const countryLocalizedName = territories[countryCode];
      const countryName =
        defaultTerritories.main.en.localeDisplayNames.territories[countryCode];
      const countryFlag = getEmojiFlag(countryCode);
      const countryCallingCode = getCountryCallingCode(countryCode);
      return {
        flagEmoji: countryFlag,
        localizedName: countryLocalizedName,
        name: countryName,
        iso2: countryCode,
        phone: countryCallingCode,
      };
    });
    const options = this._countries.map((country) => {
      return {
        triggerLabel: `${country.flagEmoji} +${country.phone}`,
        prefix: `${country.flagEmoji} +${country.phone}`,
        searchLabel: country.name,
        label: country.localizedName,
        value: country.iso2,
      };
    });

    // The detected country is not allowed.
    if (options.find((o) => o.value == initialCountry) == null) {
      initialCountry = null;
    }

    this.countryCode = initialCountry ?? options[0].value;
    this.countrySelectTarget.setAttribute(
      "data-custom-select-options-value",
      JSON.stringify(options)
    );
    this.updateValue();
  }
}
