// Type definition for intl-tel-input v18.2.1

declare type IntlTelInputAlpha2 = string;

declare interface IntlTelInputCountryData {
  name?: string;
  iso2?: IntlTelInputAlpha2;
  dialCode?: string;
}

declare type IntlTelInputPlaceholderNumberType = "MOBILE" | "FIXED_LINE";

declare interface IntlTelInputInstance {
  getExtension(): string | undefined | null;
  getNumber(): string | undefined | null;
  getNumberType(): number | undefined | null;
  getSelectedCountryData(): IntlTelInputCountryData;
  getValidationError(): number | undefined | null;
  isValidNumber(): boolean;
  isPossibleNumber(): boolean;

  destroy(): void;
  setCountry(alpha2: IntlTelInputAlpha2): void;
  setNumber(e164: string): void;
  setPlaceholderNumberType(typ: IntlTelInputPlaceholderNumberType): void;
}

declare interface IntlTelInputInitOptions {
  // UI
  autoPlaceholder?: "polite" | "aggressive" | "off";
  placeholderNumberType?: IntlTelInputPlaceholderNumberType;
  customContainer?: string;
  customPlaceholder?: (
    selectedCountryPlaceholder: string,
    selectedCountryData: IntlTelInputCountryData
  ) => string;
  dropdownContainer?: HTMLElement;

  // Functionality
  utilsScript?: string;
  nationalMode?: boolean;
  autoInsertDialCode?: boolean;
  separateDialCode?: boolean;
  allowDropdown?: boolean;
  formatOnDisplay?: boolean;
  geoIpLookup?: (
    success: (alpha2: IntlTelInputAlpha2) => void,
    failure: (err: unknown) => void
  ) => void;
  initialCountry?: IntlTelInputAlpha2;
  excludeCountries?: IntlTelInputAlpha2[];
  onlyCountries?: IntlTelInputAlpha2[];
  preferredCountries?: IntlTelInputAlpha2[];

  // Localization
  localizedCountries?: Record<IntlTelInputAlpha2, string>;

  // Form integration
  hiddenInput?: string;
}

declare interface IntlTelInputInitFunction {
  (
    element: HTMLElement,
    options?: IntlTelInputInitOptions
  ): IntlTelInputInstance;
}

declare module "intl-tel-input" {
  const intlTelInput: IntlTelInputInitFunction;
  export default intlTelInput;
}
