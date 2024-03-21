import { Controller } from "@hotwired/stimulus";

/**
 * On iOS, the keyboard is treated as overlay rather than part of the viewport.
 * This causes an extra padding to be added to the bottom of the viewport when the keyboard is open.
 *
 * This controller listens for scroll events and dismisses the keyboard when user scrolls on
 * the body or non-`overflow:auto/scroll` element.
 */
export class DismissKeyboardOnScrollController extends Controller {
  connect() {
    window.addEventListener("scroll", this.dismissKeyboard);
  }

  disconnect() {
    window.removeEventListener("scroll", this.dismissKeyboard);
  }

  dismissKeyboard = () => {
    if (
      document.activeElement instanceof HTMLElement &&
      document.activeElement === this.element
    ) {
      document.activeElement.blur();
    }
  };
}
