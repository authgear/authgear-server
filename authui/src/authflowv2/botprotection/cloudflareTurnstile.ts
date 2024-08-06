/* global Turnstile */
import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";
import {
  dispatchBotProtectionEventExpired,
  dispatchBotProtectionEventFailed,
  dispatchBotProtectionEventVerified,
} from "./botProtection";
import { parseCloudflareTurnstileErrorCode } from "./cloudflareTurnstileError";
import { setErrorMessage } from "../alert-message";

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

const CLOUDFLARE_TURNSTILE_ERROR_MSG_ID = "data-bot-protection-cloudflare";
export class CloudflareTurnstileController extends Controller {
  static values = {
    siteKey: { type: String },
    lang: { type: String },
  };

  static targets = ["widget"];

  declare siteKeyValue: string;
  declare langValue: string;
  declare widgetTarget: HTMLDivElement;
  declare widgetContainer: HTMLDivElement | undefined;
  declare widgetID: string | undefined;
  declare isReadyForRendering: boolean;

  hasExistingWidget = () => {
    return this.widgetContainer != null && this.widgetID != null;
  };
  resetWidget = () => {
    if (!this.hasExistingWidget()) {
      return;
    }
    window.turnstile.reset(this.widgetID);
  };

  renderWidget = () => {
    if (!this.isReadyForRendering) {
      throw new Error("turnstile target is not ready for rendering");
    }
    if (this.hasExistingWidget()) {
      return;
    }
    const colorScheme = getColorScheme();

    // Note how we wrap an extra layer of div here, because on cleanup we can just remove this extra layer of div.
    // container-container
    //   container <-- can just remove this on cleanup
    //     widget
    const widgetContainer = document.createElement("div");
    this.widgetContainer = widgetContainer;
    this.widgetTarget.appendChild(widgetContainer);
    const widgetID = window.turnstile.render(widgetContainer, {
      sitekey: this.siteKeyValue,
      theme: parseTheme(colorScheme),
      language: this.langValue,
      callback: (token: string) => {
        dispatchBotProtectionEventVerified(token);
      },
      "error-callback": (errCode: string) => {
        const {
          error: parsedError,
          shouldRetry,
          shouldDisplayErrMsg,
        } = parseCloudflareTurnstileErrorCode(errCode);
        console.error(
          `Cloudflare Turnstile failed with error code "${errCode}". Authgear parsed it as "${parsedError}".`
        );
        dispatchBotProtectionEventFailed(errCode);

        if (shouldRetry) {
          this.resetWidget();
        }
        if (shouldDisplayErrMsg) {
          setErrorMessage(CLOUDFLARE_TURNSTILE_ERROR_MSG_ID);
        }
        return true; // return non-falsy value to tell cloudflare we handled error already
      },
      "expired-callback": (token: string) => {
        dispatchBotProtectionEventExpired(token);
      },
      "timeout-callback": () => {
        // reset the widget to allow a visitor to solve the challenge again
        this.resetWidget();
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
    this.widgetID = widgetID ?? undefined;
  };

  undoRenderWidget = () => {
    if (this.widgetID != null) {
      window.turnstile.remove(this.widgetID);
    }
    this.widgetID = undefined;

    if (this.widgetContainer != null) {
      this.widgetTarget.removeChild(this.widgetContainer);
    }
    this.widgetContainer = undefined;
  };

  connect() {
    window.turnstile.ready(() => {
      this.isReadyForRendering = true;
    });
    document.addEventListener(
      "bot-protection-widget:render",
      this.renderWidget
    );
    document.addEventListener(
      "bot-protection-widget:undo-render",
      this.undoRenderWidget
    );
  }

  disconnect() {
    document.removeEventListener(
      "bot-protection-widget:render",
      this.renderWidget
    );
    document.removeEventListener(
      "bot-protection-widget:undo-render",
      this.undoRenderWidget
    );
    this.undoRenderWidget();
  }
}
