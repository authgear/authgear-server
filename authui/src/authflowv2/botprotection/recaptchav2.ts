/* global ReCaptchaV2 */
import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";

function parseTheme(theme: string): ReCaptchaV2.Theme | undefined {
  switch (theme) {
    case "light":
      return "light";
    case "dark":
      return "dark";
    default:
      return undefined;
  }
}

export class RecaptchaV2Controller extends Controller {
  static values = {
    siteKey: { type: String },
  };

  static targets = ["widget", "tokenInput"];

  declare siteKeyValue: string;
  declare widgetTarget: HTMLDivElement;
  declare tokenInputTargets: HTMLInputElement[];

  connect() {
    window.grecaptcha.ready(() => {
      const colorScheme = getColorScheme();
      window.grecaptcha.render(this.widgetTarget, {
        sitekey: this.siteKeyValue,
        theme: parseTheme(colorScheme),
        callback: (token: string) => {
          for (const tokenInput of this.tokenInputTargets) {
            tokenInput.value = token;
          }
        },

        // below are default values, added for clarity
        size: "normal",
      });
    });
  }
}
