import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";

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
        theme: colorScheme,
        language: this.langValue,
        callback: (token: string) => {
          for (const tokenInput of this.tokenInputTargets) {
            tokenInput.value = token;
          }
        },
        "error-callback": () => {
          // TODO: confirm handling; maybe no need to do anything?
        },
      });
      for (let i = 0; i < this.widgetTarget.children.length; i++) {
        const widget = this.widgetTarget.children[i];
        widget.classList.add("flex");
      }
    });
  }
}
