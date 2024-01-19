import { Controller } from "@hotwired/stimulus";
import zxcvbn from "zxcvbn";

enum PasswordPolicyName {
  Strength = "strength",
  Length = "length",
  Uppercase = "uppercase",
  Lowercase = "lowercase",
  Alphabet = "alphabet",
  Digit = "digit",
  Symbol = "symbol",
}

function checkPasswordLength(value: string, el: HTMLElement) {
  const minLength = Number(el.getAttribute("data-min-length"));
  const codePoints = Array.from(value);
  if (codePoints.length >= minLength) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
  }
}

function checkPasswordUppercase(value: string, el: HTMLElement) {
  if (/[A-Z]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
  }
}

function checkPasswordLowercase(value: string, el: HTMLElement) {
  if (/[a-z]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
  }
}

function checkPasswordAlphabet(value: string, el: HTMLElement) {
  if (/[a-zA-Z]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
  }
}

function checkPasswordDigit(value: string, el: HTMLElement) {
  if (/[0-9]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
  }
}

function checkPasswordSymbol(value: string, el: HTMLElement) {
  if (/[^a-zA-Z0-9]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
  }
}

function checkPasswordStrength(
  value: string,
  el: HTMLElement,
  currentMeter: HTMLMeterElement
) {
  const minLevel = Number(el.getAttribute("data-min-level"));
  const result = zxcvbn(value);
  const score = Math.min(5, Math.max(1, result.score + 1));
  currentMeter.value = score;
  if (score >= minLevel) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
  }
}

export class PasswordPolicyController extends Controller {
  static targets = ["input", "currentMeter", "policy"];

  declare inputTarget: HTMLInputElement;
  declare hasCurrentMeterTarget: boolean;
  declare currentMeterTarget: HTMLMeterElement;
  declare policyTargets: HTMLElement[];

  connect() {
    this.check();
  }

  check() {
    const value = this.inputTarget.value;
    if (value === "") {
      this.currentMeterTarget.value = -1;
      this.policyTargets.forEach((e) => {
        e.setAttribute("data-state", "");
      });

      return;
    }
    this.policyTargets.forEach((e) => {
      switch (e.getAttribute("data-password-policy-name")) {
        case PasswordPolicyName.Strength:
          if (this.hasCurrentMeterTarget) {
            checkPasswordStrength(value, e, this.currentMeterTarget);
          }
          break;
        case PasswordPolicyName.Length:
          checkPasswordLength(value, e);
          break;
        case PasswordPolicyName.Uppercase:
          checkPasswordUppercase(value, e);
          break;
        case PasswordPolicyName.Lowercase:
          checkPasswordLowercase(value, e);
          break;
        case PasswordPolicyName.Alphabet:
          checkPasswordAlphabet(value, e);
          break;
        case PasswordPolicyName.Digit:
          checkPasswordDigit(value, e);
          break;
        case PasswordPolicyName.Symbol:
          checkPasswordSymbol(value, e);
          break;
        default:
          break;
      }
    });
  }
}
