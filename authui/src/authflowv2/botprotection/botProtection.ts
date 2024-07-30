import { Controller } from "@hotwired/stimulus";
import { dispatchBotProtectionDialogOpen } from "./botProtectionDialog";

/**
 * Dispatch a custom event to set captcha verified with success token
 *
 * @param {string} token - success token
 */
export function dispatchBotProtectionEventVerified(token: string) {
  document.dispatchEvent(
    new CustomEvent("bot-protection:verify-success", {
      detail: {
        token,
      },
    })
  );
}

/**
 * Dispatch a custom event to set captcha failed
 *
 * @param {string | undefined} errMsg - error message
 */
export function dispatchBotProtectionEventFailed(errMsg?: string) {
  document.dispatchEvent(
    new CustomEvent("bot-protection:verify-failed", {
      detail: {
        errMsg,
      },
    })
  );
}

/**
 * Dispatch a custom event to set captcha expired
 *
 * @param {string | undefined} token - expired token
 */
export function dispatchBotProtectionEventExpired(token?: string) {
  document.dispatchEvent(
    new CustomEvent("bot-protection:verify-expired", {
      detail: {
        token,
      },
    })
  );
}

/**
 * Controller for bot protection verification
 *  - `verifyFormSubmit` by intercepting form submission
 *  - re-submit the intercepted form
 *
 * Expected usage:
 * - Add `data-controller="bot-protection"` to a top-level element like body
 * - Add `data-action="submit->bot-protection#verifyFormSubmit"` to any `<form>` in body
 */
export class BotProtectionController extends Controller {
  declare isVerified: boolean;
  declare formSubmitTarget: HTMLElement | null;
  verifyFormSubmit(e: Event) {
    if (!(e instanceof SubmitEvent)) {
      throw new Error("verifyFormSubmit must be triggered on submit events");
    }

    if (this.isVerified) {
      return;
    }

    e.preventDefault();
    e.stopImmediatePropagation();
    this.formSubmitTarget = e.submitter;
    dispatchBotProtectionDialogOpen();
  }

  onVerifySuccess = () => {
    this.isVerified = true;
    this.formSubmitTarget?.click();
  };

  onVerifyFailed = () => {
    this.isVerified = false;
  };

  onVerifyExpired = () => {
    this.isVerified = false;
  };

  connect() {
    document.addEventListener(
      "bot-protection:verify-success",
      this.onVerifySuccess
    );
    document.addEventListener(
      "bot-protection:verify-failed",
      this.onVerifyFailed
    );
    document.addEventListener(
      "bot-protection:verify-expired",
      this.onVerifyExpired
    );
  }
  disconnect() {
    document.removeEventListener(
      "bot-protection:verify-success",
      this.onVerifySuccess
    );
    document.removeEventListener(
      "bot-protection:verify-failed",
      this.onVerifyFailed
    );
    document.removeEventListener(
      "bot-protection:verify-expired",
      this.onVerifyExpired
    );
  }
}
