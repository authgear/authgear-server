export function setupIntlTelInput() {
  const onlyCountries = JSON.parse(document.querySelector("meta[name=x-intl-tel-input-only-countries]")?.getAttribute("content") ?? "null") ?? [];
  const preferredCountries = JSON.parse(document.querySelector("meta[name=x-intl-tel-input-preferred-countries]")?.getAttribute("content") ?? "null") ?? [];

  let initialCountry = document.querySelector("meta[name=x-geoip-country-code]")?.getAttribute("content") ?? "";
  // The detected country is not allowed.
  if (onlyCountries.indexOf(initialCountry) < 0) {
    initialCountry = "";
  }

  const elements = document.querySelectorAll("[data-intl-tel-input]");
  for (let i = 0; i < elements.length; i++) {
    const element = elements[i];
    if (element instanceof HTMLInputElement) {
      const name = element.name;
      element.name = name + "_intl_tel_input";
      const customContainer = element.getAttribute("data-intl-tel-input-class") ?? undefined;
      window.intlTelInput(element, {
        autoPlaceholder: "aggressive",
        hiddenInput: name,
        onlyCountries,
        preferredCountries,
        initialCountry,
        customContainer,
      });
    }
  }
}
