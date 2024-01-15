import { Controller } from "@hotwired/stimulus";
import zxcvbn from "zxcvbn";

function checkPasswordStrength(
  value: string,
  input: HTMLInputElement,
  currentMeter: HTMLMeterElement,
  currentMeterDescription: HTMLElement
) {
  if (value === "") {
    currentMeter.classList.add("hidden");
    return;
  }

  currentMeter.classList.remove("hidden");
  const result = zxcvbn(value);
  // FIXME(davis): Confirming with designer on the level of password strength
  const score = Math.min(5, Math.max(1, result.score + 1));
  currentMeter.value = score;
  currentMeterDescription.textContent = currentMeter.getAttribute(
    "data-desc-" + score
  );
}

export class PasswordPolicyController extends Controller {
  static targets = ["input", "currentMeter", "currentMeterDescription"];

  declare inputTarget: HTMLInputElement;
  declare currentMeterTarget: HTMLMeterElement;
  declare currentMeterDescriptionTarget: HTMLElement;

  check() {
    const value = this.inputTarget.value;
    checkPasswordStrength(
      value,
      this.inputTarget,
      this.currentMeterTarget,
      this.currentMeterDescriptionTarget
    );
  }
}
