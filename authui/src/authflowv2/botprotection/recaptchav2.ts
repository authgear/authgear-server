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
  declare readonly widgetTarget: HTMLDivElement;
  declare readonly tokenInputTarget: HTMLInputElement;

  connect() {
    window.onLoadRecaptchaV2Callback = () => {
      window.grecaptcha.render(this.widgetTarget, {
        sitekey: this.siteKeyValue,
        theme: this.themeValue,
        callback: (token: string) => {
          this.tokenInputTarget.value = token;
        },
      });
    };
  }
}
