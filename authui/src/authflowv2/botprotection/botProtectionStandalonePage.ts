import { Controller } from "@hotwired/stimulus";

/**
 * Controller for button that should be clicked on bot protection verification success
 *  - listen to captcha widget provider events
 *  - click target button
 *
 * Expected usage:
 * - Add `data-controller="bot-protection-standalone-page"` to a `<button>` element
 */
export class BotProtectionStandalonePageController extends Controller {
  getButtonElement = (): HTMLButtonElement => {
    if (!(this.element instanceof HTMLButtonElement)) {
      throw new Error(
        "bot-protection-standalone-page must be used on `<button>` elements"
      );
    }
    return this.element;
  };

  onVerifySuccess = (e: Event) => {
    if (!(e instanceof CustomEvent)) {
      throw new Error("Unexpected non-CustomEvent");
    }
    this.getButtonElement().click();
  };

  connect() {
    document.addEventListener(
      "bot-protection:verify-success",
      this.onVerifySuccess
    );
  }

  disconnect() {
    document.removeEventListener(
      "bot-protection:verify-success",
      this.onVerifySuccess
    );
  }
}
