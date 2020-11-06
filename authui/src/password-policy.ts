/// <reference path="./core.ts" />
import zxcvbn from "zxcvbn";

function checkPasswordLength(value: string, el: HTMLInputElement | null) {
  if (el == null) {
    return;
  }
  const minLength = Number(el.getAttribute("data-min-length"));
  const codePoints = Array.from(value);
  if (codePoints.length >= minLength) {
    el.classList.add("good-txt");
  }
}

function checkPasswordUppercase(value: string, el: HTMLInputElement | null) {
  if (el == null) {
    return;
  }
  if (/[A-Z]/.test(value)) {
    el.classList.add("good-txt");
  }
}

function checkPasswordLowercase(value: string, el: HTMLInputElement | null) {
  if (el == null) {
    return;
  }
  if (/[a-z]/.test(value)) {
    el.classList.add("good-txt");
  }
}

function checkPasswordDigit(value: string, el: HTMLInputElement | null) {
  if (el == null) {
    return;
  }
  if (/[0-9]/.test(value)) {
    el.classList.add("good-txt");
  }
}

function checkPasswordSymbol(value: string, el: HTMLInputElement | null) {
  if (el == null) {
    return;
  }
  if (/[^a-zA-Z0-9]/.test(value)) {
    el.classList.add("good-txt");
  }
}

function checkPasswordStrength(value: string) {
  const meter: HTMLInputElement | null = document.querySelector(
    "#password-strength-meter"
  );
  const desc: HTMLInputElement | null = document.querySelector(
    "#password-strength-meter-description"
  );
  if (meter == null || desc == null) {
    return;
  }

  meter.value = "0";
  desc.textContent = "";

  if (value === "") {
    return;
  }

  const result = zxcvbn(value);
  const score = Math.min(5, Math.max(1, result.score + 1));
  meter.value = String(score);
  desc.textContent = desc.getAttribute("data-desc-" + score);
}

window.api.onLoad(() => {
  const elems = document.querySelectorAll("[data-password-policy-password]");
  for (let i = 0; i < elems.length; i++) {
    elems[i].addEventListener("input", e => {
      const el = e.currentTarget as HTMLInputElement;
      const value = el.value;
      const els = document.querySelectorAll(".password-policy");
      for (let i = 0; i < els.length; ++i) {
        els[i].classList.remove("error-txt", "good-txt");
      }
      checkPasswordLength(
        value,
        document.querySelector(".password-policy.length")
      );
      checkPasswordUppercase(
        value,
        document.querySelector(".password-policy.uppercase")
      );
      checkPasswordLowercase(
        value,
        document.querySelector(".password-policy.lowercase")
      );
      checkPasswordDigit(
        value,
        document.querySelector(".password-policy.digit")
      );
      checkPasswordSymbol(
        value,
        document.querySelector(".password-policy.symbol")
      );
      checkPasswordStrength(value);
    });
  }
});
