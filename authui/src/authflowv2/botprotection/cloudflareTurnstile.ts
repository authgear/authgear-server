/* global Turnstile */
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
    lang: { type: String },
  };

  static targets = ["widget", "tokenInput"];

  declare siteKeyValue: string;
  declare langValue: string;
  declare widgetTarget: HTMLDivElement;
  declare tokenInputTargets: HTMLInputElement[];

  connect() {
    window.turnstile.ready(() => {
      const colorScheme = getColorScheme();
      window.turnstile.render(this.widgetTarget, {
        sitekey: this.siteKeyValue,
        theme: parseTheme(colorScheme),
        language: this.langValue,
        callback: (token: string) => {
          for (const tokenInput of this.tokenInputTargets) {
            tokenInput.value = token;
          }
        },
        size: "normal",
        "response-field": false,
      });
    });
  }
}
