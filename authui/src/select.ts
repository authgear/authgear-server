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
