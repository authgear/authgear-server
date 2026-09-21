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

const PAGE_SCROLL_LOCK_CLASS = "dialog-page-scroll-lock";

/**
 * Page scroll is locked while any dialog is open, so the backdrop placed by
 * placeOverViewport stays over the visible area. The lock is counted because
 * a dialog may open while another is still fading out.
 */
let pageScrollLockCount = 0;

function lockPageScroll(): void {
  pageScrollLockCount += 1;
  document.documentElement.classList.add(PAGE_SCROLL_LOCK_CLASS);
}

function unlockPageScroll(): void {
  pageScrollLockCount = Math.max(0, pageScrollLockCount - 1);
  if (pageScrollLockCount === 0) {
    document.documentElement.classList.remove(PAGE_SCROLL_LOCK_CLASS);
  }
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
  // Whether this dialog currently holds one count of the page scroll lock.
  private holdsScrollLock = false;

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
    this.placeOverViewport(host);
    this.acquireScrollLock();
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
    // Release the scroll lock now rather than in closeEnd: transitionend does
    // not fire when the visibility transition is skipped, and the page must
    // never stay unscrollable after the dialog is dismissed.
    this.releaseScrollLock();
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

  /**
   * The backdrop is absolutely positioned inside the host, which grows with
   * the page. Covering the whole host would centre the dialog in the
   * document, so on a long page it lands below the fold (DEV-3855). Instead,
   * cover only the part of the host that is on screen right now. Page scroll
   * is locked while the dialog is open, so this box does not need to follow
   * the scroll position.
   */
  private placeOverViewport(host: HTMLElement): void {
    if (!(this.element instanceof HTMLElement)) {
      return;
    }
    const hostRect = host.getBoundingClientRect();
    // Distance from the host's top edge down to the top of the viewport, and
    // from the bottom of the viewport down to the host's bottom edge. Both are
    // clamped at 0 so a host shorter than the viewport is still fully covered.
    const top = Math.max(0, -hostRect.top);
    const bottom = Math.max(0, hostRect.bottom - window.innerHeight);
    this.element.style.top = `${top}px`;
    this.element.style.bottom = `${bottom}px`;
  }

  private resetPlacement(): void {
    if (!(this.element instanceof HTMLElement)) {
      return;
    }
    this.element.style.removeProperty("top");
    this.element.style.removeProperty("bottom");
  }

  private acquireScrollLock(): void {
    if (this.holdsScrollLock) {
      return;
    }
    this.holdsScrollLock = true;
    lockPageScroll();
  }

  private releaseScrollLock(): void {
    if (!this.holdsScrollLock) {
      return;
    }
    this.holdsScrollLock = false;
    unlockPageScroll();
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
        this.resetPlacement();
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
    // The dialog is data-turbo-temporary: a successful submit swaps the page
    // and removes it before its closing transition can end, so make sure it
    // does not leave the page scroll locked.
    this.releaseScrollLock();
  }
}
