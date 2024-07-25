/* global Turnstile */
import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";
import {
  dispatchBotProtectionEventExpired,
  dispatchBotProtectionEventFailed,
  dispatchBotProtectionEventVerified,
} from "./botProtection";

function parseTheme(theme: string): Turnstile.Theme {
  switch (theme) {
    case "light":
      return "light";
    case "dark":
      return "dark";
    case "auto":
      return "auto";
    default:
      return "auto";
  }
}

export class CloudflareTurnstileController extends Controller {
  static values = {
    siteKey: { type: String },
    lang: { type: String },
  };

  static targets = ["widget"];

  declare siteKeyValue: string;
  declare langValue: string;
  declare widgetTarget: HTMLDivElement;

  connect() {
    window.turnstile.ready(() => {
      const colorScheme = getColorScheme();
      window.turnstile.render(this.widgetTarget, {
        sitekey: this.siteKeyValue,
        theme: parseTheme(colorScheme),
        language: this.langValue,
        callback: (token: string) => {
          dispatchBotProtectionEventVerified(token);
        },
        "error-callback": (err: string) => {
          dispatchBotProtectionEventFailed(err);

          return true; // return non-falsy value to tell cloudflare we handled error already
        },
        "expired-callback": (token: string) => {
          dispatchBotProtectionEventExpired(token);
        },
        "response-field": false,

        // below are default values, added for clarity
        size: "normal",
        appearance: "always",
        action: undefined, // no need differentiate widgets under same site key
        cData: undefined, // no need customer data, already available server-side
        "before-interactive-callback": undefined, // we do not track interactive callback
        "after-interactive-callback": undefined, // we do not track interactive callback
        "unsupported-callback": undefined, // we handled unsupported browser in error-callback by code 110500
        tabindex: 0, // a11y
        retry: "auto", // automatically retry to obtain a token on unsuccessful attempts
        "retry-interval": 8000, // default
        "refresh-expired": "auto", // automatically refresh expired token
        "refresh-timeout": "auto", // automatically refreshes upon encountering an interactive timeout

        // below fields are not available in @types/cloudflare-turnstile package yet, submitting a PR for it ref https://github.com/DefinitelyTyped/DefinitelyTyped/pull/70139
        // execution: "render", // render is default, challenge runs automatically after calling the render() function.
      });
    });
  }
}
