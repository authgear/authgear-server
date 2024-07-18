export function getColorScheme(): string {
  const queryResult = window.matchMedia("(prefers-color-scheme: dark)");
  const htmlElement = document.documentElement;
  const darkThemeEnabled =
    htmlElement.getAttribute("data-dark-theme-enabled") === "true";

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
      : implicitColorScheme;

  return colorScheme;
}
