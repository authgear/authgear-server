/**
 * This code duplicates with authui/src/colorscheme.ts. However, colorscheme.ts cannot import other modules because it is commonjs script instead
 * Please help keep code in `getColorScheme.ts` and `colorscheme.ts` sync if you are to make any changes
 */

/**
 * Get color scheme from html element
 */
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
