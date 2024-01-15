import { Controller } from "@hotwired/stimulus";

function updateMeterDescription(
  value: number,
  currentMeter: HTMLMeterElement,
  currentMeterDescription: HTMLElement
) {
  currentMeterDescription.textContent = currentMeter.getAttribute(
    "data-desc-" + value
  );
}

export class PasswordStrengthMeterController extends Controller {
  static targets = ["currentMeter", "currentMeterDescription"];

  declare currentMeterTarget: HTMLMeterElement;
  declare currentMeterDescriptionTarget: HTMLElement;

  display() {
    const value = this.currentMeterTarget.value;
    updateMeterDescription(
      value,
      this.currentMeterTarget,
      this.currentMeterDescriptionTarget
    );
  }
}
