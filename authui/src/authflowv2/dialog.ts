import { Controller } from "@hotwired/stimulus";

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
    this.element.classList.add("open");
  };

  close = () => {
    this.element.classList.remove("open");
  };

  closeOnBackgroundClick = (e: Event) => {
    if (e.target !== this.element) {
      // Clicked descendants instead of background
      return;
    }
    dispatchDialogClose(this.element.id);
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
