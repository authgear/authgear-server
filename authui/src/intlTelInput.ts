import { Controller } from "@hotwired/stimulus";

function swapElementsName(
  firstElement: HTMLInputElement,
  secondElement: HTMLInputElement
) {
  const originalName = firstElement.name;
  firstElement.name = secondElement.name;
  secondElement.name = originalName;
}

export class IntlTelInputController extends Controller {
  declare instance: IntlTelInputInstance | null;
  declare inputElement: HTMLInputElement;
  declare hiddenInputElement: HTMLInputElement;

  beforeCache = () => {
    this.inputElement.value = "";
    this.instance?.destroy();
    this.instance = null;
  };

  input() {
    this.hiddenInputElement.value =
      this.instance?.getNumber(window.intlTelInputUtils.numberFormat.E164) ??
      "";
  }

  connect() {
    if (this.instance != null) {
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

    const input = this.element as HTMLInputElement;
    const form = input.form;

    input.setAttribute("data-intl-tel-input-connecting", "true");

    // Create hidden input to the form
    const hiddenInput = document.createElement("input");
    hiddenInput.type = "hidden";
    hiddenInput.name = "x_login_id_hidden";
    form?.appendChild(hiddenInput);

    // Save the reference of input and hidden input elements
    this.inputElement = input;
    this.hiddenInputElement = hiddenInput;

    swapElementsName(input, hiddenInput);

    const customContainer =
      input.getAttribute("data-intl-tel-input-class") ?? undefined;

    this.instance = window.intlTelInput(input, {
      autoPlaceholder: "aggressive",
      onlyCountries,
      preferredCountries,
      initialCountry,
      customContainer,
    });

    input.setAttribute("data-intl-tel-input-connecting", "false");

    document.addEventListener("turbo:before-cache", this.beforeCache);
  }

  disconnect() {
    const input = this.inputElement;
    if (input.getAttribute("data-intl-tel-input-connecting")) {
      return;
    }

    const hiddenInput = this.hiddenInputElement;

    swapElementsName(input, hiddenInput);

    input.value = this.hiddenInputElement.value;

    document.removeEventListener("turbo:before-cache", this.beforeCache);
  }
}
