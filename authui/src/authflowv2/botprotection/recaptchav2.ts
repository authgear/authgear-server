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

  declare siteKeyValue: string;
  declare themeValue: string;

  connect() {
    window.onLoadRecaptchaV2Callback = () => {
      window.grecaptcha.render(this.element, {
        sitekey: this.siteKeyValue,
        theme: this.themeValue,
        callback: function (token: string) {
          alert(`Challenge Success ${token}`);
        },
      });
    };
  }
}
