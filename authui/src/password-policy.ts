import { Controller } from "@hotwired/stimulus";
import zxcvbn from "zxcvbn";

function checkPasswordLength(value: string, el: HTMLElement): boolean {
  const minLength = Number(el.getAttribute("data-min-length"));
  const codePoints = Array.from(value);
  if (codePoints.length >= minLength) {
    el.classList.add("good-txt");
    return true;
  } else {
    el.classList.add("error-txt");
    return false;
  }
}

function checkPasswordUppercase(value: string, el: HTMLElement): boolean {
  if (/[A-Z]/.test(value)) {
    el.classList.add("good-txt");
    return true;
  } else {
    el.classList.add("error-txt");
    return false;
  }
}

function checkPasswordLowercase(value: string, el: HTMLElement): boolean {
  if (/[a-z]/.test(value)) {
    el.classList.add("good-txt");
    return true;
  } else {
    el.classList.add("error-txt");
    return false;
  }
}

function checkPasswordAlphabet(value: string, el: HTMLElement): boolean {
  if (/[a-zA-Z]/.test(value)) {
    el.classList.add("good-txt");
    return true;
  } else {
    el.classList.add("error-txt");
    return false;
  }
}

function checkPasswordDigit(value: string, el: HTMLElement): boolean {
  if (/[0-9]/.test(value)) {
    el.classList.add("good-txt");
    return true;
  } else {
    el.classList.add("error-txt");
    return false;
  }
}

function checkPasswordSymbol(value: string, el: HTMLElement): boolean {
  if (/[^a-zA-Z0-9]/.test(value)) {
    el.classList.add("good-txt");
    return true;
  } else {
    el.classList.add("error-txt");
    return false;
  }
}

function getZXCVBNScore(value: string): number {
  const result = zxcvbn(value);
  const score = Math.min(5, Math.max(1, result.score + 1));
  return score;
}

function setCurrentMeterStrength(
  value: string,
  score: number,
  currentMeter: HTMLMeterElement,
  currentMeterDescription: HTMLElement
) {
  if (value === "") {
    currentMeter.value = 0;
    currentMeterDescription.textContent = "";
  } else {
    currentMeter.value = score;
    currentMeterDescription.textContent = currentMeterDescription.getAttribute(
      "data-desc-" + score
    );
  }
}

function checkPasswordStrength(
  currentMeter: HTMLMeterElement,
  requiredMeter: HTMLMeterElement,
  strengthTarget: HTMLElement
): boolean {
  if (currentMeter.value >= requiredMeter.value) {
    strengthTarget.classList.add("good-txt");
    return true;
  } else {
    strengthTarget.classList.add("error-txt");
    return false;
  }
}

export class PasswordPolicyController extends Controller {
  static targets = [
    "input",

    "submit",

    "currentMeter",
    "currentMeterDescription",

    "item",

    "length",
    "uppercase",
    "lowercase",
    "alphabet",
    "digit",
    "symbol",
    "strength",
    "requiredMeter",
  ];

  declare inputTarget: HTMLInputElement;
  declare currentMeterTarget: HTMLMeterElement;
  declare currentMeterDescriptionTarget: HTMLElement;
  declare itemTargets: HTMLElement[];

  declare hasSubmitTarget: boolean;
  declare submitTarget: HTMLElement;

  declare hasLengthTarget: boolean;
  declare lengthTarget: HTMLElement;

  declare hasUppercaseTarget: boolean;
  declare uppercaseTarget: HTMLElement;

  declare hasLowercaseTarget: boolean;
  declare lowercaseTarget: HTMLElement;

  declare hasAlphabetTarget: boolean;
  declare alphabetTarget: HTMLElement;

  declare hasDigitTarget: boolean;
  declare digitTarget: HTMLElement;

  declare hasSymbolTarget: boolean;
  declare symbolTarget: HTMLElement;

  declare hasStrengthTarget: boolean;
  declare strengthTarget: HTMLElement;

  declare hasRequiredMeterTarget: boolean;
  declare requiredMeterTarget: HTMLMeterElement;

  connect() {
    const value = this.inputTarget.value;
    if (this.hasSubmitTarget) {
      if (value === "") {
        if (
          this.submitTarget instanceof HTMLInputElement ||
          this.submitTarget instanceof HTMLButtonElement
        ) {
          this.submitTarget.disabled = true;
        }
      } else {
        this.check();
      }
    }
  }

  check() {
    const value = this.inputTarget.value;
    for (let i = 0; i < this.itemTargets.length; i++) {
      this.itemTargets[i].classList.remove("error-txt", "good-txt");
    }
    const results: boolean[] = [];
    if (this.hasLengthTarget) {
      results.push(checkPasswordLength(value, this.lengthTarget));
    }
    if (this.hasUppercaseTarget) {
      results.push(checkPasswordUppercase(value, this.uppercaseTarget));
    }
    if (this.hasLowercaseTarget) {
      results.push(checkPasswordLowercase(value, this.lowercaseTarget));
    }
    if (this.hasAlphabetTarget) {
      results.push(checkPasswordAlphabet(value, this.alphabetTarget));
    }
    if (this.hasDigitTarget) {
      results.push(checkPasswordDigit(value, this.digitTarget));
    }
    if (this.hasSymbolTarget) {
      results.push(checkPasswordSymbol(value, this.symbolTarget));
    }

    const score = getZXCVBNScore(value);
    setCurrentMeterStrength(
      value,
      score,
      this.currentMeterTarget,
      this.currentMeterDescriptionTarget
    );
    if (this.hasStrengthTarget && this.hasRequiredMeterTarget) {
      results.push(
        checkPasswordStrength(
          this.currentMeterTarget,
          this.requiredMeterTarget,
          this.strengthTarget
        )
      );
    }

    const invalid = results.some((b) => !b);
    if (this.hasSubmitTarget) {
      if (
        this.submitTarget instanceof HTMLInputElement ||
        this.submitTarget instanceof HTMLButtonElement
      ) {
        this.submitTarget.disabled = invalid;
      }
    }
  }
}
