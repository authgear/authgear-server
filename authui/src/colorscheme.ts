import { Controller } from "@hotwired/stimulus";

export class ColorSchemeController extends Controller {
  static values = {
    darkThemeEnabled: Boolean,
  };

  queryResult = window.matchMedia("(prefers-color-scheme: dark)");

  declare darkThemeEnabledValue: Boolean;

  onChange = () => {
    if (!this.darkThemeEnabledValue) {
      return;
    }

    let explicitColorScheme = "";
    const metaElement = document.querySelector("meta[name=x-color-scheme]");
    if (metaElement instanceof HTMLMetaElement) {
      explicitColorScheme = metaElement.content;
    }

    const implicitColorScheme = this.queryResult.matches ? "dark" : "light";

    const colorScheme =
      explicitColorScheme !== "" ? explicitColorScheme : implicitColorScheme;

    if (colorScheme === "dark") {
      this.element.classList.add("dark");
    } else {
      this.element.classList.remove("dark");
    }
  };

  connect() {
    this.queryResult.addEventListener("change", this.onChange);
    this.onChange();
  }

  disconnect() {
    this.queryResult.removeEventListener("change", this.onChange);
  }
}
