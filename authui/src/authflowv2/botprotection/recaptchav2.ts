/* global ReCaptchaV2 */
import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";
import {
  dispatchBotProtectionEventExpired,
  dispatchBotProtectionEventFailed,
  dispatchBotProtectionEventVerified,
} from "./botProtection";
import { setErrorMessage } from "../../setErrorMessage";

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

const RECAPTCHA_V2_ERROR_MSG_ID = "data-bot-protection-recaptcha-v2";

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
          // Tested, error-callback does not have error message/error code as params
          // Will simply fire failed event, instead of graceful handling in cloudflare
          console.error(
            "Something went wrong with Google RecaptchaV2. Please check widget for error hint."
          );
          setErrorMessage(RECAPTCHA_V2_ERROR_MSG_ID);
          dispatchBotProtectionEventFailed();
        },
        "expired-callback": () => {
          this.resetWidget();
          dispatchBotProtectionEventExpired();
        },

        // below are default values, added for clarity
        size: "normal",
        tabindex: 0,
        type: "image",
        badge: "bottomright",
      });
      this.widgetID = widgetID;
    });
  }
}
