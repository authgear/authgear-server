import { Controller } from "@hotwired/stimulus";
import zxcvbn from "zxcvbn";

function checkPasswordStrength(
  value: string,
  input: HTMLInputElement,
  currentMeter: HTMLElement,
  currentMeterDescription: HTMLElement
) {
  currentMeterDescription.textContent = "";
  var regx = new RegExp(
    "\\b" + "password-strength-meter--" + "[^ ]*[ ]?\\b",
    "g"
  );
  currentMeter.className = currentMeter.className.replace(regx, "");

  if (value === "") {
    return;
  }

  const result = zxcvbn(value);
  // FIXME(davis): Confirming with designer on the level of password strength
  const score = Math.min(5, Math.max(1, result.score + 1));
  let strengthClass = "";
  switch (score) {
    case 1:
      strengthClass = "password-strength-meter--very-weak";
      break;
    case 2:
      strengthClass = "password-strength-meter--weak";
      break;
    case 3:
      strengthClass = "password-strength-meter--medium";
      break;
    case 4:
      strengthClass = "password-strength-meter--strong";
      break;
    case 5:
      strengthClass = "password-strength-meter--very-strong";
      break;
  }
  currentMeter.classList.add(strengthClass);
  currentMeterDescription.textContent = currentMeterDescription.getAttribute(
    "data-desc-" + score
  );
}

export class PasswordPolicyController extends Controller {
  static targets = ["input", "currentMeter", "currentMeterDescription"];

  declare inputTarget: HTMLInputElement;
  declare currentMeterTarget: HTMLElement;
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
