import { Controller } from "@hotwired/stimulus";
import {
  CountryCode,
  getCountryCallingCode,
  AsYouType,
} from "libphonenumber-js";
import defaultTerritories from "cldr-localenames-full/main/en/territories.json";
import { getEmojiFlag } from "./getEmojiFlag";
import { CustomSelectController } from "./customSelect";

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

async function compileCountryList(): Promise<PhoneInputCountry[]> {
  const onlyCountryCodes = getOnlyCountryCodes();
  const preferredCountryCodes = getPreferredCountryCodes();

  const lang =
    document.querySelector("meta[name=x-locale]")?.getAttribute("content") ||
    document.documentElement.lang ||
    "en";
  const localizedTerritories = await fetch(
    `/shared-assets/cldr-localenames-full/${lang}/territories.json`
  )
    .then((r) => r.json())
    .catch(() => null);
  const territories =
    localizedTerritories?.main[lang].localeDisplayNames.territories;

  function countryCodeToCountry(countryCode: CountryCode): PhoneInputCountry {
    const countryName =
      defaultTerritories.main.en.localeDisplayNames.territories[countryCode];
    const countryLocalizedName = territories?.[countryCode] ?? countryName;
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
    const maybeCountry = asYouType.getCountry();
    if (maybeCountry) {
      this.countrySelect!.select(maybeCountry);
    }
    this.updateValue();
  }

  // countrySelect -> inputTarget.
  handleCountryInput(_event: Event): void {
    this.updateValue();
  }

  async _initPhoneCode() {
    const countries = await compileCountryList();
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
