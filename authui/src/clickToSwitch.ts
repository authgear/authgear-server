export function setupClickToSwitch(): () => void {
  const targets = document.querySelectorAll("[data-switch-to-on-click]");

  function listener(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    if (!(e.currentTarget instanceof HTMLElement)) {
      return;
    }
    const selector = e.currentTarget.getAttribute("data-switch-to-on-click");
    if (selector == null) {
      return;
    }

    const selectedElement = document.querySelector(selector);
    if (selectedElement == null) {
      return;
    }

    e.currentTarget.classList.add("hidden");
    selectedElement.classList.remove("hidden");
  }

  for (let i = 0; i < targets.length; i++) {
    const target = targets[i];
    target.addEventListener("click", listener);
  }

  return () => {
    for (let i = 0; i < targets.length; i++) {
      const target = targets[i];
      target.removeEventListener("click", listener);
    }
  };
}
