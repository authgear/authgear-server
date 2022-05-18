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

interface GenderSelect {
  selectElement: HTMLSelectElement;
  inputElement: HTMLInputElement;
  listener: (e: Event) => void;
}

export function setupGenderSelect(): () => void {
  function toggle(
    selectElement: HTMLSelectElement,
    inputElement: HTMLInputElement,
    fromListener: boolean
  ) {
    if (selectElement.value === "other") {
      inputElement.classList.remove("hidden");
      if (fromListener) {
        inputElement.value = "";
      }
    } else {
      inputElement.classList.add("hidden");
    }
  }

  const genderSelects: GenderSelect[] = [];

  const elems = document.querySelectorAll("select[data-gender-select]");
  for (let i = 0; i < elems.length; i++) {
    const selectElement = elems[i];
    if (!(selectElement instanceof HTMLSelectElement)) {
      continue;
    }
    const genderInputID = selectElement.getAttribute("data-gender-select");
    if (genderInputID == null) {
      continue;
    }
    const inputElement = document.getElementById(genderInputID);
    if (!(inputElement instanceof HTMLInputElement)) {
      continue;
    }

    const listener = () => {
      toggle(selectElement, inputElement, true);
    };

    selectElement.addEventListener("change", listener);
    toggle(selectElement, inputElement, false);

    genderSelects.push({
      selectElement,
      inputElement,
      listener,
    });
  }

  return () => {
    for (const s of genderSelects) {
      s.selectElement.removeEventListener("change", s.listener);
    }
  };
}
