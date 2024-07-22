/**
 * This code duplicates with authui/src/getColorScheme.ts. However, colorscheme.ts cannot import other modules because it is commonjs script instead
 * Please help keep code in `getColorScheme.ts` and `colorscheme.ts` sync if you are to make any changes
 */
(function () {
  const queryResult = window.matchMedia("(prefers-color-scheme: dark)");

  function onChange() {
    const htmlElement = document.documentElement;
    const darkThemeEnabled =
      htmlElement.getAttribute("data-dark-theme-enabled") === "true";
    const lightThemeEnabled =
      htmlElement.getAttribute("data-light-theme-enabled") === "true";

    let explicitColorScheme = "";
    const metaElement = document.querySelector("meta[name=x-color-scheme]");
    if (metaElement instanceof HTMLMetaElement) {
      explicitColorScheme = metaElement.content;
    }
    const queryParam = new URLSearchParams(window.location.search).get(
      "x_color_scheme"
    );
    if (queryParam != null) {
      explicitColorScheme = queryParam;
    }

    const implicitColorScheme = queryResult.matches ? "dark" : "light";

    let colorScheme = "light";
    // First of all, respect project configuration
    if (lightThemeEnabled && !darkThemeEnabled) {
      colorScheme = "light";
    } else if (!lightThemeEnabled && darkThemeEnabled) {
      colorScheme = "dark";
    } else {
      // !lightThemeEnabled && !darkThemeEnabled is treated as both enabled
      if (explicitColorScheme !== "") {
        colorScheme = explicitColorScheme;
      } else {
        colorScheme = implicitColorScheme;
      }
    }

    if (colorScheme === "dark") {
      htmlElement.classList.add("dark");
    } else {
      htmlElement.classList.remove("dark");
    }
  }

  queryResult.addEventListener("change", onChange);
  onChange();
})();
