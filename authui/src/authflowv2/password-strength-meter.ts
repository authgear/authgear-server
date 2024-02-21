import { Controller } from "@hotwired/stimulus";

function updateMeterDescription(
  currentMeter: HTMLElement,
  currentMeterDescription: HTMLElement
) {
  const ariaValueNow = currentMeter.getAttribute("aria-valuenow");
  if (!ariaValueNow) {
    return;
  }

  currentMeterDescription.textContent = currentMeter.getAttribute(
    "data-desc-" + ariaValueNow
  );
}

export class PasswordStrengthMeterController extends Controller {
  static targets = ["currentMeterDescription"];

  declare currentMeterDescriptionTarget: HTMLElement;

  observer: MutationObserver | null = null;

  connect() {
    const callback = () => {
      this.update();
    };
    this.observer = new MutationObserver(callback);
    this.observer.observe(this.element, {
      attributes: true,
    });
    this.update();
  }
  disconnect() {
    this.observer?.disconnect();
    this.observer = null;
  }

  update() {
    updateMeterDescription(
      this.element as HTMLElement,
      this.currentMeterDescriptionTarget
    );
  }
}
