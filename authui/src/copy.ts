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
  static values = {
    source: String,
    copyLabel: String,
    copiedLabel: String,
  };

  declare sourceValue: string;
  declare copyLabelValue: string;
  declare copiedLabelValue: string;
  declare hasCopyLabelValue: boolean;
  declare hasCopiedLabelValue: boolean;
  declare timeoutHandle: number | null;

  copy(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    const button = this.element as HTMLButtonElement;

    const copyLabel = this.copyLabelValue;
    const copiedLabel = this.copiedLabelValue;

    const target = document.querySelector(this.sourceValue);
    if (target == null) {
      return;
    }

    const textContent = target.textContent;
    if (textContent == null) {
      return;
    }

    copyToClipboard(textContent);

    // Show feedback
    let currentHandle = this.timeoutHandle;
    // Clear scheduled timeout if the timeout function has NOT been executed yet.
    if (currentHandle != null) {
      window.clearTimeout(currentHandle);
      this.timeoutHandle = null;
    }
    // Changing label as feedback is optional
    if (this.hasCopyLabelValue && this.hasCopiedLabelValue) {
      button.textContent = copiedLabel;
    }
    button.classList.add("outline");
    const newHandle = window.setTimeout(() => {
      // Changing label as feedback is optional
      if (this.hasCopyLabelValue && this.hasCopiedLabelValue) {
        button.textContent = copyLabel;
      }
      button.classList.remove("outline");
      this.timeoutHandle = null;
    }, 1000);
    this.timeoutHandle = newHandle;
  }
}
