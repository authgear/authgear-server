import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";

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
        theme: colorScheme,
        callback: (token: string) => {
          for (const tokenInput of this.tokenInputTargets) {
            tokenInput.value = token;
          }
          removeDefaultGRecaptchaField();
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

/**
 * This method is a workaround for Google reCaptcha v2.
 *
 * Google reCaptcha v2 inject a textarea#g-recaptcha-response field by default.
 * We do not want this field in our form, so we set it as empty string.
 *
 * Note that removing the textarea would cause error in gRecaptcha callback of
 * on solving challenge, so we opt for empty string here.
 */
function removeDefaultGRecaptchaField() {
  const gRecaptchaResponseTextArea: HTMLTextAreaElement | null =
    document.querySelector("#g-recaptcha-response");
  if (gRecaptchaResponseTextArea != null) {
    gRecaptchaResponseTextArea.value = "";
  }
}
