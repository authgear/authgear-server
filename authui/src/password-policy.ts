import { Controller } from "@hotwired/stimulus";
import zxcvbn from "zxcvbn";

function checkPasswordLength(value: string, el: HTMLElement) {
  if (el == null) {
    return;
  }
  const minLength = Number(el.getAttribute("data-min-length"));
  const codePoints = Array.from(value);
  if (codePoints.length >= minLength) {
    el.classList.add("good-txt");
  } else {
    el.classList.add("error-txt");
  }
}

function checkPasswordUppercase(value: string, el: HTMLElement) {
  if (el == null) {
    return;
  }
  if (/[A-Z]/.test(value)) {
    el.classList.add("good-txt");
  } else {
    el.classList.add("error-txt");
  }
}

function checkPasswordLowercase(value: string, el: HTMLElement) {
  if (el == null) {
    return;
  }
  if (/[a-z]/.test(value)) {
    el.classList.add("good-txt");
  } else {
    el.classList.add("error-txt");
  }
}

function checkPasswordDigit(value: string, el: HTMLElement) {
  if (el == null) {
    return;
  }
  if (/[0-9]/.test(value)) {
    el.classList.add("good-txt");
  } else {
    el.classList.add("error-txt");
  }
}

function checkPasswordSymbol(value: string, el: HTMLElement) {
  if (el == null) {
    return;
  }
  if (/[^a-zA-Z0-9]/.test(value)) {
    el.classList.add("good-txt");
  } else {
    el.classList.add("error-txt");
  }
}

function checkPasswordStrength(
  value: string,
  currentMeter: HTMLMeterElement,
  currentMeterDescription: HTMLElement,
  requiredMeter: HTMLMeterElement | null | undefined,
  strengthTarget: HTMLElement | null | undefined
) {
  currentMeter.value = 0;
  currentMeterDescription.textContent = "";

  if (value === "") {
    return;
  }

  const result = zxcvbn(value);
  const score = Math.min(5, Math.max(1, result.score + 1));
  currentMeter.value = score;
  currentMeterDescription.textContent = currentMeterDescription.getAttribute(
    "data-desc-" + score
  );

  if (requiredMeter != null && strengthTarget != null) {
    if (currentMeter.value >= requiredMeter.value) {
      strengthTarget.classList.add("good-txt");
    } else {
      strengthTarget.classList.add("error-txt");
    }
  }
}

export class PasswordPolicyController extends Controller {
  static targets = [
    "input",

    "currentMeter",
    "currentMeterDescription",

    "item",

    "length",
    "uppercase",
    "lowercase",
    "digit",
    "symbol",
    "strength",
    "requiredMeter",
  ];

  declare inputTarget: HTMLInputElement;
  declare currentMeterTarget: HTMLMeterElement;
  declare currentMeterDescriptionTarget: HTMLElement;
  declare itemTargets: HTMLElement[];

  declare hasLengthTarget: boolean;
  declare lengthTarget: HTMLElement;

  declare hasUppercaseTarget: boolean;
  declare uppercaseTarget: HTMLElement;

  declare hasLowercaseTarget: boolean;
  declare lowercaseTarget: HTMLElement;

  declare hasDigitTarget: boolean;
  declare digitTarget: HTMLElement;

  declare hasSymbolTarget: boolean;
  declare symbolTarget: HTMLElement;

  declare hasStrengthTarget: boolean;
  declare strengthTarget: HTMLElement;

  declare hasRequiredMeterTarget: boolean;
  declare requiredMeterTarget: HTMLMeterElement;

  check() {
    const value = this.inputTarget.value;
    for (let i = 0; i < this.itemTargets.length; i++) {
      this.itemTargets[i].classList.remove("error-txt", "good-txt");
    }
    if (this.hasLengthTarget) {
      checkPasswordLength(value, this.lengthTarget);
    }
    if (this.hasUppercaseTarget) {
      checkPasswordUppercase(value, this.uppercaseTarget);
    }
    if (this.hasLowercaseTarget) {
      checkPasswordLowercase(value, this.lowercaseTarget);
    }
    if (this.hasDigitTarget) {
      checkPasswordDigit(value, this.digitTarget);
    }
    if (this.hasSymbolTarget) {
      checkPasswordSymbol(value, this.symbolTarget);
    }
    if (this.hasStrengthTarget && this.hasRequiredMeterTarget) {
      checkPasswordStrength(
        value,
        this.currentMeterTarget,
        this.currentMeterDescriptionTarget,
        this.requiredMeterTarget,
        this.strengthTarget
      );
    } else {
      checkPasswordStrength(
        value,
        this.currentMeterTarget,
        this.currentMeterDescriptionTarget,
        null,
        null
      );
    }
  }
}
