import { Controller } from "@hotwired/stimulus";

/**
 * Controller for bot protection input
 *  - listen to captcha widget provider events
 *  - update form inputs for submission
 *
 * Expected usage:
 * - Add `data-controller="bot-protection-token-input"` to an `<input>` element inside `<form>` that requires bot protection verification
 */
export class BotProtectionTokenInputController extends Controller {
  getInputElement = (): HTMLInputElement => {
    if (!(this.element instanceof HTMLInputElement)) {
      throw new Error(
        "bot-protection-token-input must be used on `<input>` elements"
      );
    }
    return this.element;
  };
  onVerifySuccess = (e: Event) => {
    if (!(e instanceof CustomEvent)) {
      throw new Error("Unexpected non-CustomEvent");
    }
    this.getInputElement().value = e.detail.token;
  };

  onVerifyFailed = () => {
    this.getInputElement().value = "";
  };

  onVerifyExpired = () => {
    this.getInputElement().value = "";
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
