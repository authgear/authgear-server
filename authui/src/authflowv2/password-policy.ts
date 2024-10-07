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

interface ValidationResult {
  policy: PasswordPolicyName;
  isViolated: boolean;
}

function checkPasswordLength(value: string, el: HTMLElement): ValidationResult {
  const minLength = Number(el.getAttribute("data-min-length"));
  const codePoints = Array.from(value);
  let isViolated = false;
  if (codePoints.length >= minLength) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
    isViolated = true;
  }
  return {
    policy: PasswordPolicyName.Length,
    isViolated,
  };
}

function checkPasswordUppercase(
  value: string,
  el: HTMLElement
): ValidationResult {
  let isViolated = false;
  if (/[A-Z]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
    isViolated = true;
  }
  return {
    policy: PasswordPolicyName.Uppercase,
    isViolated,
  };
}

function checkPasswordLowercase(
  value: string,
  el: HTMLElement
): ValidationResult {
  let isViolated = false;
  if (/[a-z]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
    isViolated = true;
  }
  return {
    policy: PasswordPolicyName.Lowercase,
    isViolated,
  };
}

function checkPasswordAlphabet(value: string, el: HTMLElement) {
  let isViolated = false;
  if (/[a-zA-Z]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
    isViolated = true;
  }
  return {
    policy: PasswordPolicyName.Alphabet,
    isViolated,
  };
}

function checkPasswordDigit(value: string, el: HTMLElement) {
  let isViolated = false;
  if (/[0-9]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
    isViolated = true;
  }
  return {
    policy: PasswordPolicyName.Digit,
    isViolated,
  };
}

function checkPasswordSymbol(value: string, el: HTMLElement) {
  let isViolated = false;
  if (/[^a-zA-Z0-9]/.test(value)) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
    isViolated = true;
  }
  return {
    policy: PasswordPolicyName.Symbol,
    isViolated,
  };
}

function checkPasswordStrength(
  value: string,
  el: HTMLElement,
  currentMeter: HTMLElement
) {
  let isViolated = false;
  const minLevel = Number(el.getAttribute("data-min-level"));
  const result = zxcvbn(value);
  const score = Math.min(5, Math.max(1, result.score + 1));
  currentMeter.setAttribute("aria-valuenow", String(score));
  if (score >= minLevel) {
    el.setAttribute("data-state", "pass");
  } else {
    el.setAttribute("data-state", "fail");
    isViolated = true;
  }
  return {
    policy: PasswordPolicyName.Strength,
    isViolated,
  };
}

export class PasswordPolicyController extends Controller {
  static targets = ["input", "currentMeter", "policy"];
  static ATTR_POLICY_VIOLATED = "data-password-policy-violated";

  declare inputTarget: HTMLInputElement;
  declare hasCurrentMeterTarget: boolean;
  declare currentMeterTarget: HTMLElement;
  declare policyTargets: HTMLElement[];

  connect() {
    void this.check();
  }

  async check() {
    const value = this.inputTarget.value;
    if (value === "") {
      if (this.hasCurrentMeterTarget) {
        this.currentMeterTarget.setAttribute("aria-valuenow", "0");
      }
      this.policyTargets.forEach((e) => {
        e.setAttribute("data-state", "");
      });

      return;
    }
    const violatedPolicies: PasswordPolicyName[] = [];
    // eslint-disable-next-line sonarjs/cognitive-complexity,complexity
    this.policyTargets.forEach((e) => {
      switch (e.getAttribute("data-password-policy-name")) {
        case PasswordPolicyName.Strength:
          if (this.hasCurrentMeterTarget) {
            const result = checkPasswordStrength(
              value,
              e,
              this.currentMeterTarget
            );
            if (result.isViolated) {
              violatedPolicies.push(result.policy);
            }
          }
          break;
        case PasswordPolicyName.Length: {
          const result = checkPasswordLength(value, e);
          if (result.isViolated) {
            violatedPolicies.push(result.policy);
          }
          break;
        }
        case PasswordPolicyName.Uppercase: {
          const result = checkPasswordUppercase(value, e);
          if (result.isViolated) {
            violatedPolicies.push(result.policy);
          }
          break;
        }
        case PasswordPolicyName.Lowercase: {
          const result = checkPasswordLowercase(value, e);
          if (result.isViolated) {
            violatedPolicies.push(result.policy);
          }
          break;
        }
        case PasswordPolicyName.Alphabet: {
          const result = checkPasswordAlphabet(value, e);
          if (result.isViolated) {
            violatedPolicies.push(result.policy);
          }
          break;
        }
        case PasswordPolicyName.Digit: {
          const result = checkPasswordDigit(value, e);
          if (result.isViolated) {
            violatedPolicies.push(result.policy);
          }
          break;
        }
        case PasswordPolicyName.Symbol: {
          const result = checkPasswordSymbol(value, e);
          if (result.isViolated) {
            violatedPolicies.push(result.policy);
          }
          break;
        }
        default:
          break;
      }
    });
    if (violatedPolicies.length > 0) {
      this.inputTarget.setAttribute(
        PasswordPolicyController.ATTR_POLICY_VIOLATED,
        violatedPolicies.join(" ")
      );
    } else {
      this.inputTarget.removeAttribute(
        PasswordPolicyController.ATTR_POLICY_VIOLATED
      );
    }
  }
}
