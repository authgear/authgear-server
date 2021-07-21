export function setupIntlTelInput() {
  const onlyCountries = JSON.parse(document.querySelector("meta[name=x-intl-tel-input-only-countries]")?.getAttribute("content") ?? "null") ?? [];
  const preferredCountries = JSON.parse(document.querySelector("meta[name=x-intl-tel-input-preferred-countries]")?.getAttribute("content") ?? "null") ?? [];

  let initialCountry = document.querySelector("meta[name=x-geoip-country-code]")?.getAttribute("content") ?? "";
  // The detected country is not allowed.
  if (onlyCountries.indexOf(initialCountry) < 0) {
    initialCountry = "";
  }

  const instances: (ReturnType<typeof window.intlTelInput>)[] = [];
  const elements = document.querySelectorAll("[data-intl-tel-input]");
  for (let i = 0; i < elements.length; i++) {
    const element = elements[i];
    if (element instanceof HTMLInputElement) {
      // Store the original form field name.
      let originalName = element.getAttribute("data-intl-tel-input-name");
      if (originalName == null || originalName === "") {
        originalName = element.name;
      }

      // Rename the name of this form field,
      // because the actual input being used is hiddenInput.
      element.name = originalName + "_intl_tel_input";

      const customContainer = element.getAttribute("data-intl-tel-input-class") ?? undefined;
      const instance = window.intlTelInput(element, {
        hiddenInput: originalName,
        autoPlaceholder: "aggressive",
        onlyCountries,
        preferredCountries,
        initialCountry,
        customContainer,
      });
      instances.push(instance);
    }
  }

  return () => {
    for (const instance of instances) {
      instance.destroy();
    }
  };
}
