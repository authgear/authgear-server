(function () {
  const queryResult = window.matchMedia("(prefers-color-scheme: dark)");

  function onChange() {
    const htmlElement = document.documentElement;
    const darkThemeEnabled =
      htmlElement.getAttribute("data-dark-theme-enabled") === "true";
    const darkThemeForced =
      htmlElement.getAttribute("data-dark-theme-forced") === "true";

    let explicitColorScheme = "";
    const metaElement = document.querySelector("meta[name=x-color-scheme]");
    if (metaElement instanceof HTMLMetaElement) {
      explicitColorScheme = metaElement.content;
    }

    const implicitColorScheme = queryResult.matches ? "dark" : "light";

    const colorScheme = !darkThemeEnabled
      ? "light"
      : explicitColorScheme !== ""
        ? explicitColorScheme
        : darkThemeForced
          ? "dark"
          : implicitColorScheme;

    if (colorScheme === "dark") {
      htmlElement.classList.add("dark");
    } else {
      htmlElement.classList.remove("dark");
    }
  }

  queryResult.addEventListener("change", onChange);
  onChange();
})();
