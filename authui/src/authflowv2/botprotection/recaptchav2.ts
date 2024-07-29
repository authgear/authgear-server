/* global ReCaptchaV2 */
import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";
import {
  dispatchBotProtectionEventExpired,
  dispatchBotProtectionEventFailed,
  dispatchBotProtectionEventVerified,
} from "./botProtection";
import { setErrorMessage } from "../alert-message";

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
  declare widgetID: number | undefined;

  hasExistingWidget = () => {
    return this.widgetTarget.innerHTML !== "" && this.widgetID != null;
  };
  resetWidget = () => {
    if (!this.hasExistingWidget()) {
      return;
    }
    window.grecaptcha.reset(this.widgetID); // default to first widget created
  };

  connect() {
    window.grecaptcha.ready(() => {
      if (this.hasExistingWidget()) {
        return;
      }
      const colorScheme = getColorScheme();

      // Note how we wrap an extra layer of div here, because on cleanup we can just remove this extra layer of div.
      // container-container
      //   container <-- can just remove this on cleanup
      //     widget
      const widgetContainer = document.createElement("div");
      this.widgetTarget.appendChild(widgetContainer);
      const widgetID = window.grecaptcha.render(widgetContainer, {
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

  disconnect() {
    this.widgetID = undefined;
    this.widgetTarget.innerHTML = ""; // Remove widget container
  }
}
