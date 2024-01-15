import { Controller } from "@hotwired/stimulus";
import zxcvbn from "zxcvbn";

function checkPasswordStrength(value: string, currentMeter: HTMLMeterElement) {
  if (value === "") {
    currentMeter.value = -1;
    return;
  }

  currentMeter.classList.remove("hidden");
  const result = zxcvbn(value);
  const score = Math.min(5, Math.max(1, result.score + 1));
  currentMeter.value = score;
}

export class PasswordPolicyController extends Controller {
  static targets = ["input", "currentMeter"];

  declare inputTarget: HTMLInputElement;
  declare currentMeterTarget: HTMLMeterElement;

  check() {
    const value = this.inputTarget.value;
    checkPasswordStrength(value, this.currentMeterTarget);
    const event = new CustomEvent("password-strength-updated");
    window.dispatchEvent(event);
  }
}
