export function setupSelectEmptyValue(): () => void {
  function toggleClass(elem: HTMLSelectElement) {
    if (elem.value === "") {
      elem.classList.add("empty");
    } else {
      elem.classList.remove("empty");
    }
  }

  function listener(e: Event) {
    if (e.target instanceof HTMLSelectElement) {
      toggleClass(e.target);
    }
  }

  const selects = document.querySelectorAll("select");

  for (let i = 0; i < selects.length; i++) {
    const selectElem = selects[i];
    selectElem.addEventListener("change", listener);
    toggleClass(selectElem);
  }

  return () => {
    for (let i = 0; i < selects.length; i++) {
      const selectElem = selects[i];
      selectElem.removeEventListener("change", listener);
    }
  };
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
