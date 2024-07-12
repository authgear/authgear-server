import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";

function parseTheme(theme: string): Turnstile.Theme {
  switch (theme) {
    case "light":
      return "light";
    case "dark":
      return "dark";
    case "auto":
      return "auto";
    default:
      return "auto";
  }
}

export class CloudflareTurnstileController extends Controller {
  static values = {
    siteKey: { type: String },
  };

  static targets = ["widget", "tokenInput"];

  declare siteKeyValue: string;
  declare widgetTarget: HTMLDivElement;
  declare tokenInputTargets: HTMLInputElement[];

  connect() {
    window.turnstile.ready(() => {
      const _theme: Turnstile.Theme = "auto";
      const colorScheme = getColorScheme();
      window.turnstile.render(this.widgetTarget, {
        sitekey: this.siteKeyValue,
        theme: parseTheme(colorScheme),
        callback: (token: string) => {
          for (const tokenInput of this.tokenInputTargets) {
            tokenInput.value = token;
          }
        },
      });
      for (let i = 0; i < this.widgetTarget.children.length; i++) {
        const widget = this.widgetTarget.children[i];
        widget.classList.add("flex");
      }
    });
  }
}
