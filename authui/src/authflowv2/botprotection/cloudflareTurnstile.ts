import { Controller } from "@hotwired/stimulus";

declare global {
  interface Window {
    turnstile: any;
  }
}

export class CloudflareTurnstileController extends Controller {
  static values = {
    siteKey: { type: String },
    theme: { type: String },
  };

  declare siteKeyValue: string;
  declare themeValue: string;

  connect() {
    window.turnstile.ready(() => {
      window.turnstile.render(this.element, {
        sitekey: this.siteKeyValue,
        theme: this.themeValue,
        callback: function (token: string) {
          alert(`Challenge Success ${token}`);
        },
      });
    });
  }
}
