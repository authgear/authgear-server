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

  static targets = ["widget", "tokenInput"];

  declare siteKeyValue: string;
  declare themeValue: string;
  declare widgetTarget: HTMLDivElement;
  declare tokenInputTarget: HTMLInputElement;

  connect() {
    window.turnstile.ready(() => {
      window.turnstile.render(this.widgetTarget, {
        sitekey: this.siteKeyValue,
        theme: this.themeValue,
        callback: (token: string) => {
          this.tokenInputTarget.value = token;
        },
      });
    });
  }
}
