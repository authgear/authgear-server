import { Controller } from "@hotwired/stimulus";

const CLOSE_ANIMATION_DURATION_MS = 280;

/**
 * Dispatch a custom event to set target dialog open
 */
export function dispatchDialogOpen(dialogID: string) {
  document.dispatchEvent(new CustomEvent(`dialog-${dialogID}:open`));
}

/**
 * Dispatch a custom event to set target dialog close
 */
export function dispatchDialogClose(dialogID: string) {
  document.dispatchEvent(new CustomEvent(`dialog-${dialogID}:close`));
}

/**
 * Controller for dialog display
 *
 * Expected usage:
 * - Add `data-controller="dialog"` to a dialog
 * - Specific `id` attribute to that HTML element
 */
export class DialogController extends Controller {
  open = () => {
    this.element.classList.remove("close");
    this.element.classList.add("open");
  };

  close = () => {
    this.element.classList.add("close");
    // We want close animation to show before `display: none` is applied
    setTimeout(
      () => this.element.classList.remove("open"),
      CLOSE_ANIMATION_DURATION_MS
    );
  };

  connect() {
    document.addEventListener(`dialog-${this.element.id}:open`, this.open);
    document.addEventListener(`dialog-${this.element.id}:close`, this.close);
  }

  disconnect() {
    document.removeEventListener(`dialog-${this.element.id}:open`, this.open);
    document.removeEventListener(`dialog-${this.element.id}:close`, this.close);
  }
}
