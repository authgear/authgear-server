import { Controller } from "@hotwired/stimulus";

/**
 * Dispatch a custom event to set target dialog open
 */
export function dispatchDialogOpen(dialogID: string) {
  document.dispatchEvent(
    new CustomEvent(`dialog:open`, { detail: { id: dialogID } })
  );
}

/**
 * Dispatch a custom event to set target dialog close
 */
export function dispatchDialogClose(dialogID: string) {
  document.dispatchEvent(
    new CustomEvent(`dialog:close`, { detail: { id: dialogID } })
  );
}

/**
 * Dispatch a custom event to publish target dialog open event
 */
function dispatchDialogOpenEnd(dialogID: string) {
  document.dispatchEvent(
    new CustomEvent(`dialog:opened`, { detail: { id: dialogID } })
  );
}

/**
 * Dispatch a custom event to publish target dialog close event
 */
function dispatchDialogCloseEnd(dialogID: string) {
  document.dispatchEvent(
    new CustomEvent(`dialog:closed`, { detail: { id: dialogID } })
  );
}

/**
 * Controller for dialog display
 *
 * Expected usage:
 * - Add `data-controller="dialog"` to a dialog
 * - Specific `id` attribute to that HTML element
 *
 * @listens dialog:open
 * @listens dialog:close
 * @fires dialog:opened
 * @fires dialog:closed
 *
 * @example // To open a dialog, dispatch below event
 *     new CustomEvent("dialog:open", {detail: {id: "foobar"}})
 * @example // To close a dialog, dispatch below event
 *     new CustomEvent("dialog:close", {detail: {id: "foobar"}})
 * @example // To receive a callback when the dialog is opened, listen to following event
 *     new CustomEvent("dialog:opened", {detail: {id: "foobar"}})
 * @example // To receive a callback when the dialog is closed, listen to following event
 *     new CustomEvent("dialog:closed", {detail: {id: "foobar"}})
 */
export class DialogController extends Controller {
  open = (e: Event) => {
    if (!(e instanceof CustomEvent)) {
      return;
    }
    if (this.element.id !== e.detail.id) {
      // open event targets other dialog
      return;
    }
    this.element.classList.add("open");
  };

  close = (e: Event) => {
    if (!(e instanceof CustomEvent)) {
      return;
    }
    if (this.element.id !== e.detail.id) {
      // close event targets other dialog
      return;
    }
    this.element.classList.remove("open");
  };

  get isOpened() {
    return this.element.classList.contains("open");
  }

  get isClosed() {
    return !this.isOpened;
  }

  openEnd = (e: Event) => {
    const isVisibilityEvent =
      (e as TransitionEvent).propertyName === "visibility";
    if (isVisibilityEvent && this.isOpened) {
      dispatchDialogOpenEnd(this.element.id);
    }
  };

  closeEnd = (e: Event) => {
    const isVisibilityEvent =
      (e as TransitionEvent).propertyName === "visibility";

    if (isVisibilityEvent && this.isClosed) {
      dispatchDialogCloseEnd(this.element.id);
    }
  };

  closeOnCrossBtnClick = () => {
    dispatchDialogClose(this.element.id);
  };

  closeOnBackgroundClick = (e: Event) => {
    if (e.target !== this.element) {
      // Clicked descendants instead of background
      return;
    }
    dispatchDialogClose(this.element.id);
  };

  connect() {
    document.addEventListener(`dialog:open`, this.open);
    document.addEventListener(`dialog:close`, this.close);
    this.element.addEventListener("transitionend", this.openEnd);
    this.element.addEventListener("transitionend", this.closeEnd);
  }

  disconnect() {
    document.removeEventListener(`dialog:open`, this.open);
    document.removeEventListener(`dialog:close`, this.close);
    this.element.removeEventListener("transitionend", this.openEnd);
    this.element.removeEventListener("transitionend", this.closeEnd);
  }
}
