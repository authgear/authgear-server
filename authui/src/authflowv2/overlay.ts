import { Controller } from "@hotwired/stimulus";

export class OverlayController extends Controller {
  static values = { defaultOpen: Boolean };

  declare readonly defaultOpenValue: boolean;

  connect(): void {
    this.element.addEventListener("transitionstart", this.transitionstart);
    this.element.addEventListener("transitionend", this.transitionend);
    if (this.defaultOpenValue) {
      this.open();
    }
  }

  disconnect(): void {
    this.element.removeEventListener("transitionstart", this.transitionstart);
    this.element.removeEventListener("transitionend", this.transitionend);
  }

  transitionstart = (e: Event) => {
    if (e instanceof TransitionEvent) {
      // This is intentionally empty as we have nothing to do here.
    }
  };

  transitionend = (e: Event) => {
    if (e instanceof TransitionEvent) {
      const host = this.getHost();
      if (host != null) {
        this.revertPrepareHost(host);
      }
    }
  };

  private getHost(): HTMLElement | null {
    const dialogHost = getComputedStyle(
      document.documentElement
    ).getPropertyValue("--overlay-host");

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

  open() {
    const host = this.getHost();
    if (host == null) {
      return;
    }

    this.prepareHost(host);
    this.element.classList.add("open");
  }

  close() {
    this.element.classList.remove("open");
  }
}
