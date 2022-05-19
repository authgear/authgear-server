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
  static targets = ["input", "hiddenInput"];

  declare inputTarget: HTMLInputElement;
  declare hiddenInputTarget: HTMLInputElement;
  declare instance: IntlTelInputInstance | null;

  input() {
    this.hiddenInputTarget.value =
      this.instance?.getNumber(window.intlTelInputUtils.numberFormat.E164) ??
      "";
  }

  connect() {
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

    const input = this.inputTarget;
    const hiddenInput = this.hiddenInputTarget;

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
  }

  disconnect() {
    const input = this.inputTarget;
    const hiddenInput = this.hiddenInputTarget;

    swapElementsName(input, hiddenInput);

    input.value = this.hiddenInputTarget.value;

    this.instance?.destroy();
    this.instance = null;
  }
}
