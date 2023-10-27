import { Controller } from "@hotwired/stimulus";

function mutationObserverIsAvailable(): boolean {
  return typeof window.MutationObserver !== "undefined";
}

// Mirror the state of a button and transfer click event.
export class MirrorButtonController extends Controller {
  static values = {
    selector: String,
  };

  declare selectorValue: string;

  observer: MutationObserver | null = null;
  target: HTMLButtonElement | null = null;

  connect() {
    if (mutationObserverIsAvailable()) {
      const target = document.querySelector(this.selectorValue);
      if (target instanceof HTMLButtonElement) {
        this.target = target;
        const callback = (
          mutationRecords: MutationRecord[],
          _observer: MutationObserver
        ) => {
          for (const mutationRecord of mutationRecords) {
            if (mutationRecord.type === "attributes") {
              this.mirror();
            }
          }
        };
        this.observer = new MutationObserver(callback);
        this.observer.observe(target, {
          attributes: true,
        });
        this.mirror();
      }
    }
  }

  disconnect() {
    this.observer?.disconnect();
    this.observer = null;
  }

  mirror() {
    const element = this.element;
    const target = this.target;
    if (element instanceof HTMLButtonElement && target != null) {
      element.disabled = target.disabled;
    }
  }

  click(e: Event) {
    const target = this.target;
    if (target != null) {
      e.preventDefault();
      target.click();
    }
  }
}
