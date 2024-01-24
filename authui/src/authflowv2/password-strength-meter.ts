import { Controller } from "@hotwired/stimulus";

function updateMeterDescription(
  currentMeter: HTMLMeterElement,
  currentMeterDescription: HTMLElement
) {
  currentMeterDescription.textContent = currentMeter.getAttribute(
    "data-desc-" + currentMeter.value
  );
}

export class PasswordStrengthMeterController extends Controller {
  static targets = ["currentMeterDescription"];

  declare currentMeterDescriptionTarget: HTMLElement;

  observer: MutationObserver | null = null;

  connect() {
    const callback = () => {
      this.display();
    };
    this.observer = new MutationObserver(callback);
    this.observer.observe(this.element, {
      attributes: true,
    });
  }
  disconnect() {
    this.observer?.disconnect();
    this.observer = null;
  }

  display() {
    updateMeterDescription(
      this.element as HTMLMeterElement,
      this.currentMeterDescriptionTarget
    );
  }
}
