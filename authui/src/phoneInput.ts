import { Controller } from "@hotwired/stimulus";
import {
  CountryCode,
  getCountryCallingCode,
  AsYouType,
} from "libphonenumber-js";
import defaultTerritories from "cldr-localenames-full/main/en/territories.json";
import territoriesMap from "cldr-localenames-full/main/*/territories.json";
import { getEmojiFlag } from "./getEmojiFlag";
import { CustomSelectController } from "./customSelect";
import metadata from "libphonenumber-js/metadata.min.json";
const CountryCodes = metadata.country_calling_codes;
// Duplicated country codes
// {
//     "1": [ "US", "AG", "AI", "AS", "BB", "BM", "BS", "CA", "DM", "DO", "GD", "GU", "JM", "KN", "KY", "LC", "MP", "MS", "PR", "SX", "TC", "TT", "VC", "VG", "VI" ],
//     "7": [ "RU", "KZ" ],
//     "39": [ "IT", "VA" ],
//     "44": [ "GB", "GG", "IM", "JE" ],
//     "47": [ "NO", "SJ" ],
//     "61": [ "AU", "CC", "CX" ],
//     "212": [ "MA", "EH" ],
//     "262": [ "RE", "YT" ],
//     "290": [ "SH", "TA" ],
//     "358": [ "FI", "AX" ],
//     "590": [ "GP", "BL", "MF" ],
//     "599": [ "CW", "BQ" ]
// }

const defaultCountryForDuplicatedCountryCodes: Record<string, CountryCode> = {
  "1": "US",
  "7": "RU",
  "39": "IT",
  "44": "GB",
  "47": "NO",
  "61": "AU",
  "212": "MA",
  "262": "RE",
  "290": "SH",
  "358": "FI",
  "590": "GP",
  "599": "CW",
};

interface PhoneInputCountry {
  flagEmoji: string;
  localizedName: string;
  name: string;
  iso2: string;
  phone: string;
}

function getOnlyCountryCodes(): CountryCode[] {
  const onlyCountries: CountryCode[] =
    JSON.parse(
      document
        .querySelector("meta[name=x-phone-input-only-countries]")
        ?.getAttribute("content") ?? "null"
    ) ?? [];
  return onlyCountries;
}

function getPreferredCountryCodes(): CountryCode[] {
  const preferredCountries: CountryCode[] =
    JSON.parse(
      document
        .querySelector("meta[name=x-phone-input-preferred-countries]")
        ?.getAttribute("content") ?? "null"
    ) ?? [];
  return preferredCountries;
}

function compileCountryList(): PhoneInputCountry[] {
  const onlyCountryCodes = getOnlyCountryCodes();
  const preferredCountryCodes = getPreferredCountryCodes();

  const lang = document.documentElement.lang || "en";
  const localizedTerritories = territoriesMap[lang];
  const territories =
    localizedTerritories?.main[lang as keyof typeof localizedTerritories.main]
      ?.localeDisplayNames.territories ||
    defaultTerritories.main.en.localeDisplayNames.territories;

  function countryCodeToCountry(countryCode: CountryCode): PhoneInputCountry {
    const countryLocalizedName = territories[countryCode];
    const countryName =
      defaultTerritories.main.en.localeDisplayNames.territories[countryCode];
    const countryFlag = getEmojiFlag(countryCode);
    const countryCallingCode = getCountryCallingCode(countryCode);
    return {
      flagEmoji: `<span class="country-flag-icon phone-input__country-flag twemoji-countries">${countryFlag}</span>`,
      localizedName: countryLocalizedName,
      name: countryName,
      iso2: countryCode,
      phone: countryCallingCode,
    };
  }

  const onlyCountries = onlyCountryCodes.map(countryCodeToCountry);
  const preferredCountries = preferredCountryCodes.map(countryCodeToCountry);

  onlyCountries.sort((a, b) => {
    return a.localizedName.localeCompare(b.localizedName);
  });

  const countries = [];
  const seen = new Set();

  for (const c of preferredCountries) {
    if (seen.has(c.iso2)) {
      continue;
    }
    seen.add(c.iso2);
    countries.push(c);
  }
  for (const c of onlyCountries) {
    if (seen.has(c.iso2)) {
      continue;
    }
    seen.add(c.iso2);
    countries.push(c);
  }
  return countries;
}

export class PhoneInputController extends Controller {
  static targets = ["countrySelect", "input", "phoneInput"];

  declare readonly countrySelectTarget: HTMLElement;
  declare readonly inputTarget: HTMLInputElement;
  declare readonly phoneInputTarget: HTMLInputElement;

  get countrySelect(): CustomSelectController | null {
    const ctr = this.application.getControllerForElementAndIdentifier(
      this.countrySelectTarget,
      "custom-select"
    );
    return ctr as CustomSelectController | null;
  }

  set value(newValue: string) {
    this.inputTarget.value = newValue;
  }

  // countrySelect, phoneInputTarget -> inputTarget
  updateValue(): void {
    const countryValue = this.countrySelect?.value;
    const rawValue = this.phoneInputTarget.value;

    if (rawValue.startsWith("+")) {
      this.value = rawValue;
    } else if (countryValue != null) {
      const newValue = `+${getCountryCallingCode(
        countryValue as CountryCode
      )}${rawValue}`;
      this.value = newValue;
    } else {
      this.value = rawValue;
    }
  }

  decomposeValue(
    value: string
  ): [countryCode: CountryCode | null, remainings: string] {
    const asYouType = new AsYouType();
    asYouType.input(value);
    let inputValue = value;
    const countryCode = asYouType.getCountry() ?? null;
    if (countryCode != null) {
      const callingCode = "+" + getCountryCallingCode(countryCode);
      inputValue = value.replace(callingCode, "");
    }
    return [countryCode, inputValue];
  }

  // phoneInputTarget -> countrySelect AND inputTarget.
  handleNumberInput(event: Event): void {
    const target = event.target as HTMLInputElement;
    let value = target.value;
    const asYouType = new AsYouType();
    asYouType.input(value);

    let countryCodeFromPartialNumber: CountryCode | undefined = undefined;
    const callingCode = asYouType.getCallingCode();
    if (callingCode != null) {
      countryCodeFromPartialNumber =
        defaultCountryForDuplicatedCountryCodes[callingCode] ??
        CountryCodes[callingCode][0];
    }
    const maybeCountry = asYouType.getCountry();
    if (maybeCountry || countryCodeFromPartialNumber) {
      this.countrySelect!.select(maybeCountry ?? countryCodeFromPartialNumber);
    }
    this.updateValue();
  }

  // countrySelect -> inputTarget.
  handleCountryInput(_event: Event): void {
    this.updateValue();
  }

  async _initPhoneCode() {
    const countries = compileCountryList();
    const options = countries.map((country) => {
      return {
        triggerLabel: `${country.flagEmoji} +${country.phone}`,
        prefix: `${country.flagEmoji} +${country.phone}`,
        searchLabel: country.name,
        label: country.localizedName,
        value: country.iso2,
      };
    });
    this.countrySelectTarget.setAttribute(
      "data-custom-select-options-value",
      JSON.stringify(options)
    );

    // 1. If this.inputTarget.value has something.
    // 2. x-geoip-country-code.
    // 3. If countryCode is invalid, reset to empty.
    // 4. Select the first one if countryCode is empty.

    const geoIPCountryCode: CountryCode | null =
      (document
        .querySelector("meta[name=x-geoip-country-code]")
        ?.getAttribute("content") as CountryCode) ?? null;

    let countryCode: CountryCode | null = null;
    let inputValue: string = this.phoneInputTarget.value;

    if (this.inputTarget.value !== "") {
      [countryCode, inputValue] = this.decomposeValue(this.inputTarget.value);
      this.phoneInputTarget.value = inputValue;
    }

    // countryCode is still null.
    if (countryCode == null && geoIPCountryCode != null) {
      countryCode = geoIPCountryCode;
    }

    // The detected country is not allowed.
    if (options.find((o) => o.value == countryCode) == null) {
      countryCode = null;
    }

    const initialValue = countryCode ?? options[0].value;
    this.setCountrySelectValue(initialValue);
  }

  connect() {
    this._initPhoneCode();

    window.addEventListener("pageshow", this.handlePageShow);
  }

  disconnect() {
    window.removeEventListener("pageshow", this.handlePageShow);
  }

  inputTargetConnected() {
    this._initPhoneCode();
  }

  handlePageShow = () => {
    // Restore the value from bfcache
    const restoredValue = this.inputTarget.value;
    if (!restoredValue) {
      return;
    }

    const [countryCode, inputValue] = this.decomposeValue(
      this.inputTarget.value
    );
    this.phoneInputTarget.value = inputValue;

    if (countryCode != null) {
      this.setCountrySelectValue(countryCode);
    }
  };

  private setCountrySelectValue(newValue: string) {
    if (this.countrySelect != null) {
      this.countrySelect.select(newValue);
    } else {
      this.countrySelectTarget.setAttribute(
        "data-custom-select-initial-value-value",
        newValue
      );
    }
  }
}
