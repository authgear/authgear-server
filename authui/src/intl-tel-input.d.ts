// Type definition for intl-tel-input v17.0.13

declare module "intl-tel-input";

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
  useFullscreenPopup?: boolean;

  // Functionality
  utilsScript?: string;
  nationalMode?: boolean;
  autoHideDialCode?: boolean;
  separateDialCode?: boolean;
  allowDropdown?: boolean;
  formatOnDisplay?: boolean;
  geoIpLookup?: (
    success: (alpha2: IntlTelInputAlpha2) => void,
    failure: (err: unknown) => void
  ) => void;
  initialCountry?: "auto" | IntlTelInputAlpha2;
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

declare interface IntlTelInputGlobals {
  getCountryData(): IntlTelInputCountryData[];
  getInstance(element: HTMLElement): IntlTelInputInstance | null | undefined;
  loadUtils(s: string): Promise<void>;
}

// Inject things into window.
declare interface Window {
  intlTelInputGlobals: IntlTelInputGlobals;
  intlTelInput: IntlTelInputInitFunction;
  intlTelInputUtils: {
    numberFormat: {
      E164: number;
    };
  };
}
