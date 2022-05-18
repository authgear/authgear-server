import { Controller } from "@hotwired/stimulus";

export class SelectEmptyValueController extends Controller {
  static targets = ["select"];

  declare selectTarget: HTMLSelectElement;

  toggleClass() {
    const selectElement = this.selectTarget;

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

  toggle({ params }: { params?: { fromlistener: boolean } }) {
    const selectElement = this.selectTarget;
    const inputElement = this.inputTarget;
    const fromListener: boolean = params?.fromlistener ?? false;

    if (selectElement.value === "other") {
      inputElement.classList.remove("hidden");
      if (fromListener) {
        inputElement.value = "";
      }
    } else {
      inputElement.classList.add("hidden");
    }
  }

  connect() {
    // passing empty object to indicate this operation is not calling from listener
    this.toggle({});
  }
}
