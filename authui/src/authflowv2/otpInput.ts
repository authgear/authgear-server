import { Controller } from "@hotwired/stimulus";

export class OtpInputController extends Controller {
  static targets = ["input", "submit", "digitsContainer"];

  declare readonly inputTarget: HTMLInputElement;
  declare readonly submitTarget: HTMLButtonElement;
  declare readonly digitsContainerTarget: HTMLElement;

  spans: HTMLElement[] = [];

  get maxLength(): number {
    if (this.inputTarget.maxLength) {
      return this.inputTarget.maxLength;
    }

    return 6;
  }

  get value(): string {
    return this.inputTarget.value;
  }

  beforeCache = () => {
    this.inputTarget.value = "";
  };

  connect(): void {
    this.inputTarget.classList.remove("input");
    this.inputTarget.classList.add("with-js");
    this.inputTarget.addEventListener("input", this.handleInput);
    this.inputTarget.addEventListener("paste", this.handlePaste);
    this.inputTarget.addEventListener("focus", this.handleFocus);
    this.inputTarget.addEventListener("blur", this.handleBlur);
    this.submitTarget.disabled = true;
    // element.selectionchange is NOT the same as document.selectionchange
    // element.selectionchange is an experimental technology.
    window.document.addEventListener(
      "selectionchange",
      this.handleSelectionChange
    );
    document.addEventListener("turbo:before-cache", this.beforeCache);
    this.render();
  }

  disconnect(): void {
    this.inputTarget.removeEventListener("input", this.handleInput);
    this.inputTarget.removeEventListener("paste", this.handlePaste);
    this.inputTarget.removeEventListener("focus", this.handleFocus);
    this.inputTarget.removeEventListener("blur", this.handleBlur);
    window.document.removeEventListener(
      "selectionchange",
      this.handleSelectionChange
    );
    document.removeEventListener("turbo:before-cache", this.beforeCache);
  }

  _setValue = (value: string): void => {
    this.inputTarget.value = value
      .replace(/[^0-9]/g, "")
      .slice(0, this.maxLength);

    const reachedMaxDigits = this.value.length === this.maxLength;
    if (reachedMaxDigits) {
      this.submitTarget.disabled = false;
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

  handleFocus = (event: FocusEvent): void => {
    const input = event.target as HTMLInputElement;
    input.setSelectionRange(input.value.length, input.value.length);
    this.render();
  };

  handleBlur = (): void => {
    this.render();
  };

  handleSelectionChange = (_event: Event): void => {
    if (this.inputTarget === document.activeElement) {
      this.inputTarget.setSelectionRange(
        this.inputTarget.value.length,
        this.inputTarget.value.length
      );
    }
  };

  isSpanSelected = (index: number): boolean => {
    const isFocused = this.inputTarget === document.activeElement;
    const isNextBox = this.value.length === index;
    return isFocused && isNextBox;
  };

  render = (): void => {
    const digitsContainer = this.digitsContainerTarget;
    if (this.spans.length !== this.maxLength) {
      digitsContainer.innerHTML = "";
    }

    for (let i = 0; i < this.maxLength; i++) {
      let textContent = this.value.slice(i, i + 1) || "";
      const classes = this.isSpanSelected(i)
        ? ["otp-input__digit", "otp-input__digit--focus"]
        : ["otp-input__digit"];

      const isLastDigit = i < this.value.length - 1;
      const isBlurred = this.inputTarget !== document.activeElement;
      if (textContent && (isLastDigit || isBlurred)) {
        textContent = " ";
        classes.push("otp-input__digit--masked");
      }

      this.inputTarget.style.letterSpacing = `calc(${this.inputTarget.offsetWidth}px / ${this.maxLength})`;

      let span = this.spans[i];
      // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
      if (!span) {
        span = document.createElement("span");
        digitsContainer.appendChild(span);
        this.spans[i] = span;
      }

      span.textContent = textContent;
      span.className = classes.join(" ");
    }
  };
}
