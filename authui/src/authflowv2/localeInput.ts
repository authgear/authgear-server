import { Controller } from "@hotwired/stimulus";
import { CustomSelectController } from "./customSelect";

export class LocaleInputController extends Controller {
  static targets = ["localeSelect", "input", "localeSelectInput", "disable"];

  declare readonly localeSelectTarget: HTMLElement;
  declare readonly inputTarget: HTMLInputElement;
  declare readonly localeSelectInputTarget: HTMLInputElement;
  declare readonly disableTarget: HTMLElement;

  get localeSelect(): CustomSelectController | null {
    const ctr = this.application.getControllerForElementAndIdentifier(
      this.localeSelectTarget,
      "custom-select"
    );
    return ctr as CustomSelectController | null;
  }

  get value(): string {
    return this.inputTarget.value;
  }

  set value(newValue: string) {
    this.inputTarget.value = newValue;
  }

  private isReady: boolean = false;

  updateValue(): void {
    const localeValue = this.localeSelect?.value;
    this.value = localeValue ?? "";

    this.setSubmittable();
  }

  handleLocaleInput(_event: Event): void {
    this.updateValue();
  }

  private async initLocaleValue() {
    this.isReady = true;

    this.setLocaleSelectValue(this.value);

    this.setSubmittable();
  }

  setSubmittable() {
    if (this.value) {
      this.disableTarget.toggleAttribute("disabled", false);
    } else {
      this.disableTarget.toggleAttribute("disabled", true);
    }
  }

  connect() {
    void this.initLocaleValue();
    this.localeSelectTarget.classList.remove("hidden");
    this.inputTarget.classList.add("hidden");

    window.addEventListener("pageshow", this.handlePageShow);
  }

  disconnect() {
    window.removeEventListener("pageshow", this.handlePageShow);
  }

  handlePageShow = () => {
    // Restore the value from bfcache
    const restoredValue = this.inputTarget.value;
    if (!restoredValue) {
      return;
    }

    if (this.inputTarget.value) {
      this.setLocaleSelectValue(this.inputTarget.value);
    }
  };

  handleInputBlur = () => {
    if (this.inputTarget.value) {
      this.setLocaleSelectValue(this.inputTarget.value);
    }
  };

  private setLocaleSelectValue(newValue: string) {
    if (!this.isReady) {
      return;
    }
    if (this.localeSelect != null) {
      this.localeSelect.select(newValue);
    } else {
      this.localeSelectTarget.setAttribute(
        "data-custom-select-initial-value-value",
        newValue
      );
    }
  }
}
