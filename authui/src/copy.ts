import { Controller } from "@hotwired/stimulus";

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

export class CopyButtonController extends Controller {
  static targets = ["source", "button"];

  declare sourceTarget: HTMLElement;
  declare buttonTarget: HTMLButtonElement;

  copy(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    const button = this.buttonTarget;

    const copyLabel = button.getAttribute("data-copy-button-copy-label");
    const copiedLabel = button.getAttribute("data-copy-button-copied-label");

    const target = this.sourceTarget;
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
}
