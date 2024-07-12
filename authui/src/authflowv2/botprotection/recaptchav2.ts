import { Controller } from "@hotwired/stimulus";

declare global {
  interface Window {
    grecaptcha: any;
    onLoadRecaptchaV2Callback: any;
  }
}

export class RecaptchaV2Controller extends Controller {
  static values = {
    siteKey: { type: String },
    theme: { type: String },
  };

  static targets = ["widget", "tokenInput"];

  declare siteKeyValue: string;
  declare themeValue: string;
  declare widgetTarget: HTMLDivElement;
  declare tokenInputTargets: HTMLInputElement[];

  connect() {
    window.onLoadRecaptchaV2Callback = () => {
      window.grecaptcha.render(this.widgetTarget, {
        sitekey: this.siteKeyValue,
        theme: this.themeValue,
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
    };
  }
}
