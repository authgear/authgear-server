import { Controller } from "@hotwired/stimulus";

function swapElementsName(
  firstElement: HTMLInputElement,
  secondElement: HTMLInputElement
) {
  const originalName = firstElement.name;
  firstElement.name = secondElement.name;
  secondElement.name = originalName;
}

function buildE164Value(countryCallingCode: string, rawValue: string): string {
  const trimmedValue = rawValue.replace(/[^+0-9]/g, "");
  const prefix = trimmedValue.slice(0, countryCallingCode.length);
  if (prefix === countryCallingCode) {
    return trimmedValue;
  }
  return `${countryCallingCode}${trimmedValue}`;
}

export class IntlTelInputController extends Controller {
  static values = {
    class: String,
  };

  declare classValue: string;
  declare instance: IntlTelInputInstance | null;
  declare hiddenInputElement: HTMLInputElement;

  // When we call window.intlTelInput,
  // disconnect will be called immediately, followed by a connect.
  // We have to detect this situation, otherwise we will stick in a infinite loop.
  // The initialization sequence looks like connect,disconnect,connect
  // Therefore, the first time we enter connect, we expect 1 disconnect and 1 connect to be followed.
  // Here we use two booleans to keep track of that.
  ignoreConnect: boolean = false;
  ignoreDisconnect: boolean = false;

  input(fromConnect: boolean) {
    const instance = this.instance;
    if (instance == null) {
      return;
    }

    // If it is a valid number, then getNumber() returns a E.164 number already.
    // Otherwise we build a value that is a prefix of a E164 number.
    const isValid = instance.isPossibleNumber();
    if (isValid) {
      const s = instance.getNumber();
      if (typeof s === "string") {
        this.hiddenInputElement.value = s;
        // Emit the custom event "input" to cooperate with RetainFormFormController
        this.dispatch("input", {
          detail: {
            value: s,
          },
        });
      }
    } else {
      const { iso2, dialCode } = instance.getSelectedCountryData();
      if (fromConnect === true && iso2 != null && dialCode != null) {
        const countryCallingCode = `+${dialCode}`;
        this.inputElement.value = this.inputElement.value.replace(
          countryCallingCode,
          ""
        );
        instance.setCountry(iso2);
      }

      if (dialCode != null) {
        const countryCallingCode = `+${dialCode}`;
        const value = this.inputElement.value;
        const s = buildE164Value(countryCallingCode, value);
        this.hiddenInputElement.value = s;
        // Emit the custom event "input" to cooperate with RetainFormFormController
        this.dispatch("input", {
          detail: {
            value: s,
          },
        });
      }
    }
  }

  get inputElement(): HTMLInputElement {
    return this.element as HTMLInputElement;
  }

  connect() {
    if (this.ignoreConnect) {
      this.ignoreConnect = false;
      return;
    }

    const onlyCountries =
      JSON.parse(
        document
          .querySelector("meta[name=x-intl-tel-input-only-countries]")
          ?.getAttribute("content") ?? "null"
      ) ?? [];

    const preferredCountries =
      JSON.parse(
        document
          .querySelector("meta[name=x-intl-tel-input-preferred-countries]")
          ?.getAttribute("content") ?? "null"
      ) ?? [];

    let initialCountry =
      document
        .querySelector("meta[name=x-geoip-country-code]")
        ?.getAttribute("content") ?? "";
    // The detected country is not allowed.
    if (onlyCountries.indexOf(initialCountry) < 0) {
      initialCountry = "";
    }

    const form = this.inputElement.form;

    // Create hidden input to the form
    const hiddenInput = document.createElement("input");
    hiddenInput.type = "hidden";
    hiddenInput.name = "x_login_id_hidden";
    form?.appendChild(hiddenInput);

    // Save the reference of input and hidden input elements
    this.hiddenInputElement = hiddenInput;

    swapElementsName(this.inputElement, hiddenInput);

    const customContainer = this.classValue;

    this.ignoreConnect = true;
    this.ignoreDisconnect = true;
    this.instance = window.intlTelInput(this.inputElement, {
      autoPlaceholder: "aggressive",
      onlyCountries,
      preferredCountries,
      initialCountry: initialCountry.toLowerCase(),
      customContainer,
      useFullscreenPopup: false,
    });

    this.input(true);
  }

  disconnect() {
    if (this.ignoreDisconnect) {
      this.ignoreDisconnect = false;
      return;
    }

    // When we disconnect, we store the value in the value attribute.
    // Next time when we connect again, the value is used to initialize the input.
    const input = this.inputElement;
    const value = this.hiddenInputElement.value;
    input.setAttribute("value", value);
    input.value = value;
    const hiddenInput = this.hiddenInputElement;
    swapElementsName(input, hiddenInput);
    hiddenInput.parentNode?.removeChild(hiddenInput);

    this.instance?.destroy();
    this.instance = null;
  }
}
