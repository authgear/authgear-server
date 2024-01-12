import { Controller } from "@hotwired/stimulus";

export class OtpInputController extends Controller {
  static targets = ["input", "submit", "digitsContainer"];

  declare readonly inputTarget: HTMLInputElement;
  declare readonly submitTarget: HTMLButtonElement;
  declare readonly digitsContainerTarget: HTMLElement;

  value: string = "";
  maxLength: number = 6;
  masked: boolean = true;

  spans: HTMLElement[] = [];

  connect(): void {
    this.inputTarget.addEventListener("input", this.handleInput);
    this.inputTarget.addEventListener("paste", this.handlePaste);
    this.inputTarget.addEventListener("focus", this.handleFocus);
    this.inputTarget.addEventListener("blur", this.handleBlur);
    this.inputTarget.addEventListener("keyup", this.handleKeyUp);
    this.inputTarget.addEventListener("selectionchange", this.handleSelect);

    // Set initial value and render
    this.value = this.inputTarget.value || "";
    this.maxLength = this.inputTarget.maxLength || 6;
    this.render();
  }

  disconnect(): void {
    this.inputTarget.removeEventListener("input", this.handleInput);
    this.inputTarget.removeEventListener("paste", this.handlePaste);
    this.inputTarget.removeEventListener("focus", this.handleFocus);
    this.inputTarget.removeEventListener("blur", this.handleBlur);
    this.inputTarget.removeEventListener("keyup", this.handleKeyUp);
    this.inputTarget.removeEventListener("selectionchange", this.handleSelect);
  }

  _setValue = (value: string): void => {
    this.inputTarget.value = value
      .replace(/[^0-9]/g, "")
      .slice(0, this.maxLength);
    this.value = this.inputTarget.value;

    const reachedMaxDigits = this.value.length === this.maxLength;
    if (reachedMaxDigits) {
      this.submitTarget.click();
    }

    this.render();
  };

  handleInput = (event: Event): void => {
    const input = event.target as HTMLInputElement;
    this._setValue(input.value);
  };

  handlePaste = (event: ClipboardEvent): void => {
    event.preventDefault();
    const text = event.clipboardData?.getData("text/plain");
    if (text) {
      this._setValue(text);
    }
  };

  handleFocus = (): void => {
    this.render();
  };

  handleBlur = (): void => {
    this.render();
  };

  handleSelect = (): void => {
    this.render();
  };

  handleKeyUp = (event: KeyboardEvent): void => {
    if (event.key === "Backspace") {
      event.preventDefault();
      this._setValue("");
    }
  };

  isSpanSelected = (index: number): boolean => {
    const isFocused = this.inputTarget === document.activeElement;
    const caretStart = this.inputTarget.selectionStart || 0;
    const caretEnd = this.inputTarget.selectionEnd || 1;

    return isFocused && index + 1 >= caretStart && index < caretEnd;
  };

  render = (): void => {
    const digitsContainer = this.digitsContainerTarget;
    if (this.spans.length !== this.maxLength) {
      digitsContainer.innerHTML = "";
    }

    for (let i = 0; i < this.maxLength; i++) {
      let textContent = this.value.slice(i, i + 1) || "";
      let className = this.isSpanSelected(i)
        ? "otp-input__digit otp-input__digit--focus"
        : "otp-input__digit";

      if (this.masked && textContent !== "") {
        textContent = " ";
        className += " otp-input__digit--masked";
      }

      this.inputTarget.style.fontSize = `1px`;
      this.inputTarget.style.letterSpacing = `calc(${this.inputTarget.offsetWidth}px / ${this.maxLength})`;

      let span = this.spans[i];
      if (!span) {
        span = document.createElement("span");
        digitsContainer.appendChild(span);
        this.spans[i] = span;
      }

      span.textContent = textContent;
      span.className = className;
    }
  };
}
