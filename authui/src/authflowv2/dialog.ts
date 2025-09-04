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
function dispatchDialogOpenStart(dialogID: string) {
  document.dispatchEvent(
    new CustomEvent(`dialog:open-start`, { detail: { id: dialogID } })
  );
}

/**
 * Dispatch a custom event to publish target dialog close event
 */
function dispatchDialogCloseEnd(dialogID: string) {
  document.dispatchEvent(
    new CustomEvent(`dialog:close-end`, { detail: { id: dialogID } })
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
  open() {
    dispatchDialogOpen(this.element.id);
  }

  close() {
    dispatchDialogClose(this.element.id);
  }

  private openFromEvent = (e: Event) => {
    if (!(e instanceof CustomEvent)) {
      return;
    }
    if (this.element.id !== e.detail.id) {
      // open event targets other dialog
      return;
    }

    const host = this.getHost();
    if (host == null) {
      return;
    }

    this.prepareHost(host);
    this.element.classList.add("open");
    const activeElement = document.activeElement;
    if (activeElement instanceof HTMLElement) {
      activeElement.blur();
    }
  };

  private closeFromEvent = (e: Event) => {
    if (!(e instanceof CustomEvent)) {
      return;
    }
    if (this.element.id !== e.detail.id) {
      // close event targets other dialog
      return;
    }
    this.element.classList.remove("open");
  };

  private getHost(): HTMLElement | null {
    const dialogHost = getComputedStyle(
      document.documentElement
    ).getPropertyValue("--dialog-host");

    const host = document.querySelector(dialogHost);
    if (host == null) {
      return null;
    }
    if (host instanceof HTMLElement) {
      return host;
    }
    return null;
  }

  private prepareHost(host: HTMLElement): void {
    host.classList.add("relative");
  }

  private revertPrepareHost(host: HTMLElement): void {
    host.classList.remove("relative");
  }

  get isOpened() {
    return this.element.classList.contains("open");
  }

  get isClosed() {
    return !this.isOpened;
  }

  openStart = (e: Event) => {
    if (e instanceof TransitionEvent) {
      const isVisibilityEvent = e.propertyName === "visibility";
      if (isVisibilityEvent && this.isOpened) {
        dispatchDialogOpenStart(this.element.id);
      }
    }
  };

  closeEnd = (e: Event) => {
    if (e instanceof TransitionEvent) {
      const isVisibilityEvent = e.propertyName === "visibility";
      if (isVisibilityEvent && this.isClosed) {
        const host = this.getHost();
        if (host != null) {
          this.revertPrepareHost(host);
        }
        dispatchDialogCloseEnd(this.element.id);
      }
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
    document.addEventListener(`dialog:open`, this.openFromEvent);
    document.addEventListener(`dialog:close`, this.closeFromEvent);
    this.element.addEventListener("transitionstart", this.openStart);
    this.element.addEventListener("transitionend", this.closeEnd);
  }

  disconnect() {
    document.removeEventListener(`dialog:open`, this.openFromEvent);
    document.removeEventListener(`dialog:close`, this.closeFromEvent);
    this.element.removeEventListener("transitionstart", this.openStart);
    this.element.removeEventListener("transitionend", this.closeEnd);
  }
}
