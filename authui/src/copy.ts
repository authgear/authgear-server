function copyToClipboard(str: string): void {
  const el = document.createElement("textarea");
  el.value = str;
  // Set non-editable to avoid focus and move outside of view
  el.setAttribute("readonly", "");
  el.setAttribute("style", "position: absolute; left: -9999px");
  document.body.appendChild(el);
  // Select text inside element
  el.select();
  el.setSelectionRange(0, el.value.length); // for mobile device
  document.execCommand("copy");
  // Remove temporary element
  document.body.removeChild(el);
}

export function setupCopyButton(): () => void {
  function copy(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    const button = e.currentTarget as HTMLElement;
    const targetSelector = button.getAttribute("data-copy-button-target");
    if (targetSelector == null) {
      return;
    }

    const copyLabel = button.getAttribute("data-copy-button-copy-label");
    const copiedLabel = button.getAttribute("data-copy-button-copied-label");

    const target = document.querySelector(targetSelector);
    if (target == null) {
      return;
    }

    const textContent = target.textContent;
    if (textContent == null) {
      return;
    }

    copyToClipboard(textContent);

    // Show feedback
    const currentHandle = button.getAttribute(
      "data-copy-button-timeout-handle"
    );
    // Clear scheduled timeout if the timeout function has NOT been executed yet.
    if (currentHandle != null) {
      window.clearTimeout(Number(currentHandle));
      button.removeAttribute("data-copy-button-timeout-handle");
    }
    // Changing label as feedback is optional
    if (copyLabel != null && copiedLabel != null) {
      button.textContent = copiedLabel;
    }
    button.classList.add("outline");
    const newHandle = window.setTimeout(() => {
      // Changing label as feedback is optional
      if (copyLabel != null && copiedLabel != null) {
        button.textContent = copyLabel;
      }
      button.classList.remove("outline");
      button.removeAttribute("data-copy-button-timeout-handle");
    }, 1000);
    button.setAttribute("data-copy-button-timeout-handle", String(newHandle));
  }

  const elems = document.querySelectorAll("[data-copy-button-target]");
  const buttons: HTMLElement[] = [];
  for (let i = 0; i < elems.length; i++) {
    const elem = elems[i];
    if (elem instanceof HTMLElement) {
      buttons.push(elem);
    }
  }
  for (const button of buttons) {
    button.addEventListener("click", copy);
  }
  return () => {
    for (const button of buttons) {
      button.removeEventListener("click", copy);
    }
  };
}
