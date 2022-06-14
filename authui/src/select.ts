import { Controller } from "@hotwired/stimulus";

export class SelectEmptyValueController extends Controller {
  toggleClass() {
    const selectElement = this.element as HTMLSelectElement;

    if (selectElement.value === "") {
      selectElement.classList.add("empty");
    } else {
      selectElement.classList.remove("empty");
    }
  }

  connect() {
    this.toggleClass();
  }
}

export class GenderSelectController extends Controller {
  static targets = ["select", "input"];

  declare selectTarget: HTMLSelectElement;
  declare inputTarget: HTMLInputElement;

  toggle(fromListener: boolean) {
    const selectElement = this.selectTarget;
    const inputElement = this.inputTarget;

    if (selectElement.value === "other") {
      inputElement.classList.remove("hidden");
      if (fromListener) {
        inputElement.value = "";
      }
    } else {
      inputElement.classList.add("hidden");
    }
  }

  onChange() {
    this.toggle(true);
  }

  connect() {
    this.toggle(false);
  }
}
