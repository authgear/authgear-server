/* global ReCaptchaV2 */
import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";
import {
  dispatchBotProtectionEventExpired,
  dispatchBotProtectionEventFailed,
  dispatchBotProtectionEventVerified,
} from "./botProtection";

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

  static targets = ["widget"];

  declare siteKeyValue: string;
  declare widgetTarget: HTMLDivElement;
  declare widgetID: number;

  resetWidget = () => {
    window.grecaptcha.reset(this.widgetID); // default to first widget created
  };

  connect() {
    window.grecaptcha.ready(() => {
      const colorScheme = getColorScheme();
      const widgetID = window.grecaptcha.render(this.widgetTarget, {
        sitekey: this.siteKeyValue,
        theme: parseTheme(colorScheme),
        callback: (token: string) => {
          dispatchBotProtectionEventVerified(token);
        },
        "error-callback": () => {
          dispatchBotProtectionEventFailed();
        },
        "expired-callback": () => {
          this.resetWidget();
          dispatchBotProtectionEventExpired();
        },

        // below are default values, added for clarity
        size: "normal",
      });
      this.widgetID = widgetID;
    });
  }
}
