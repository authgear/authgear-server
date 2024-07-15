import { Controller } from "@hotwired/stimulus";
import {
  CountryCode,
  getCountryCallingCode,
  AsYouType,
  default as parsePhoneNumber,
} from "libphonenumber-js";
import defaultTerritories from "cldr-localenames-full/main/en/territories.json";
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

const defaultCountryForDuplicatedCountryCodes: Record<
  string,
  CountryCode | undefined
> = {
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

type Territories =
  typeof defaultTerritories.main.en.localeDisplayNames.territories;

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

async function compileDefaultCountryList() {
  return compileCountryList(
    defaultTerritories.main.en.localeDisplayNames.territories
  );
}

async function compileLocalizedCountryList() {
  const lang =
    (document
      .querySelector("meta[name=x-cldr-locale]")
      ?.getAttribute("content") ??
      "") ||
    document.documentElement.lang ||
    "en";
  const localizedTerritories = await fetch(
    `/shared-assets/cldr-localenames-full/${lang}/territories.json`
  )
    .then(async (r) => r.json())
    .catch(() => null);
  const territories =
    localizedTerritories?.main[lang].localeDisplayNames.territories;

  if (territories == null) {
    return compileDefaultCountryList();
  }

  return compileCountryList(territories);
}

async function compileCountryList(
  territories: Territories
): Promise<PhoneInputCountry[]> {
  const onlyCountryCodes = getOnlyCountryCodes();
  const preferredCountryCodes = getPreferredCountryCodes();

  function countryCodeToCountry(countryCode: CountryCode): PhoneInputCountry {
    const countryName =
      defaultTerritories.main.en.localeDisplayNames.territories[countryCode];
    const countryLocalizedName = territories[countryCode] || countryName;
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
  static targets = [
    "countrySelect",
    "input",
    "phoneInput",
    "countrySelectInput",
  ];

  declare readonly countrySelectTarget: HTMLElement;
  declare readonly inputTarget: HTMLInputElement;
  declare readonly phoneInputTarget: HTMLInputElement;
  declare readonly countrySelectInputTarget: HTMLInputElement;

  get countrySelect(): CustomSelectController | null {
    const ctr = this.application.getControllerForElementAndIdentifier(
      this.countrySelectTarget,
      "custom-select"
    );
    return ctr as CustomSelectController | null;
  }

  get value(): string {
    return this.inputTarget.value;
  }

  set value(newValue: string) {
    this.inputTarget.value = newValue;
  }

  private isReady: boolean = false;

  // countrySelect, phoneInputTarget -> inputTarget
  updateValue(): void {
    const countryValue = this.countrySelect?.value;
    const rawValue = this.phoneInputTarget.value;
    let combinedValue: string = rawValue;

    if (rawValue.startsWith("+")) {
      combinedValue = rawValue;
    } else if (countryValue != null) {
      combinedValue = `+${getCountryCallingCode(
        countryValue as CountryCode
      )}${rawValue}`;
    }

    const parsed = parsePhoneNumber(combinedValue);
    if (parsed != null) {
      combinedValue = parsed.format("E.164");
    }

    this.value = combinedValue;
  }

  decomposeValue(
    value: string
  ): [countryCode: CountryCode | null, remainings: string] {
    const asYouType = new AsYouType();
    asYouType.input(value);
    let inputValue = value;
    let countryCodeFromPartialNumber: CountryCode | undefined;
    const callingCode = asYouType.getCallingCode();
    if (callingCode != null) {
      countryCodeFromPartialNumber =
        defaultCountryForDuplicatedCountryCodes[callingCode] ??
        CountryCodes[callingCode][0];
    }

    let countryCode = asYouType.getCountry() ?? null;

    // Determine country code in the following order
    // 1. asYouType.getCountry()
    // 2. asYouType.getCallingCode() and map with defaultCountryForDuplicatedCountryCodes
    if (!countryCode && countryCodeFromPartialNumber) {
      countryCode = countryCodeFromPartialNumber;
    }

    if (countryCode) {
      const callingCode = "+" + getCountryCallingCode(countryCode);
      inputValue = value.replace(callingCode, "");
    }

    return [countryCode, inputValue];
  }

  // phoneInputTarget -> countrySelect AND inputTarget.
  handleNumberInput(event: Event): void {
    const target = event.target as HTMLInputElement;
    const value = target.value;
    const [maybeCountry, _remainings] = this.decomposeValue(value);
    if (maybeCountry) {
      this.setCountrySelectValue(maybeCountry);
    }
    this.updateValue();
  }

  // countrySelect -> inputTarget.
  handleCountryInput(_event: Event): void {
    this.updateValue();
  }

  private async initPhoneCode() {
    const countryListSources = [
      compileDefaultCountryList,
      compileLocalizedCountryList,
    ];

    for (const source of countryListSources) {
      try {
        const countries = await source();
        const options = countries.map((country) => {
          return {
            triggerLabel: `${country.flagEmoji} +${country.phone}`,
            prefix: `${country.flagEmoji} +${country.phone}`,
            searchLabel: `+${country.phone} ${country.name}`,
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
            ?.getAttribute("content") as CountryCode | undefined) ?? null;

        let countryCode: CountryCode | null = null;
        let inputValue: string = this.phoneInputTarget.value;

        if (this.inputTarget.value !== "") {
          [countryCode, inputValue] = this.decomposeValue(
            this.inputTarget.value
          );
          this.phoneInputTarget.value = inputValue;
        }

        this.isReady = true;

        // countryCode is still null.
        if (countryCode == null && geoIPCountryCode != null) {
          countryCode = geoIPCountryCode;
        }

        // The detected country is not allowed.
        if (options.find((o) => o.value === countryCode) == null) {
          countryCode = null;
        }

        const initialValue = countryCode ?? options[0].value;
        this.setCountrySelectValue(initialValue);
      } catch (e: unknown) {
        console.error(e);
      }
    }
  }

  connect() {
    void this.initPhoneCode();
    this.phoneInputTarget.addEventListener("blur", this.handleInputBlur);
    this.phoneInputTarget.parentElement?.classList.remove("hidden");
    this.countrySelectInputTarget.parentElement?.classList.remove("hidden");
    this.inputTarget.classList.add("hidden");

    window.addEventListener("pageshow", this.handlePageShow);
  }

  disconnect() {
    this.phoneInputTarget.removeEventListener("blur", this.handleInputBlur);
    window.removeEventListener("pageshow", this.handlePageShow);
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

  handleInputBlur = () => {
    const [maybeCountry, remainings] = this.decomposeValue(
      this.inputTarget.value
    );
    if (maybeCountry) {
      this.setCountrySelectValue(maybeCountry);
      this.phoneInputTarget.value = remainings;
    }
  };

  private setCountrySelectValue(newValue: string) {
    if (!this.isReady) {
      return;
    }
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
