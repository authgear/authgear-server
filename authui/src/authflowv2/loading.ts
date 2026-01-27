import { Controller } from "@hotwired/stimulus";

interface LoadingHandle {
  onError: () => void;
  onFinally: () => void;
}

function disableAllButtons(): () => void {
  const buttons = document.querySelectorAll("button");
  const original: [HTMLButtonElement, boolean][] = [];
  for (let i = 0; i < buttons.length; i++) {
    const button = buttons[i];
    const disabled = button.disabled;
    const state: [HTMLButtonElement, boolean] = [button, disabled];
    button.disabled = true;
    original.push(state);
  }

  return () => {
    for (let i = 0; i < original.length; i++) {
      const [button, disabled] = original[i];
      button.disabled = disabled;
    }
  };
}

function makeAllButtonPointerEventsNone() {
  const buttons = document.querySelectorAll("button");
  for (let i = 0; i < buttons.length; i++) {
    const button = buttons[i];
    // We cannot simply use button.disabled = true because
    // in the event handler of "submit", disabling a button will cause it
    // to be excluded from the form body.
    //
    // Therefore, we can only change our stylesheet to apply disabled style to [data-disabled] as well.
    button.setAttribute("data-disabled", "true");
    button.classList.add("pointer-events-none");
  }
}

export class LoadingController extends Controller {
  revert: (() => void) | null = null;

  // Revert disabled buttons before Turbo caches the page
  // To avoid flickering in the UI
  beforeCache = () => {
    this.revert?.();
    this.revert = null;
  };

  onSubmit = (e: SubmitEvent) => {
    // We do not call preventDefault() nor stopPropgation() here.
    // We want the browser to submit the form.
    // We just want to make the buttons no longer react to pointer events.

    const form = e.target;
    if (form instanceof HTMLFormElement) {
      const turboFormController =
        this.application.getControllerForElementAndIdentifier(
          form,
          "turbo-form"
        );
      // Detect if the form is controlled by turbo-form.
      // We only step in if the form is NOT controlled by turbo-form.
      // We also do not step in if the form is submitted to a new tab.
      if (turboFormController == null && form.target !== "_blank") {
        makeAllButtonPointerEventsNone();
      }
    }
  };

  startTurboFormSubmission(e: HTMLElement | null): LoadingHandle {
    this.revert = disableAllButtons();
    e?.setAttribute("data-loading", "true");

    return {
      onError: () => {
        this.revert?.();
        this.revert = null;
        e?.removeAttribute("data-loading");
      },
      onFinally: () => {
        e?.removeAttribute("data-loading");
      },
    };
  }

  connect() {
    document.addEventListener("turbo:before-cache", this.beforeCache);
    document.addEventListener("submit", this.onSubmit);
  }

  disconnect() {
    document.removeEventListener("turbo:before-cache", this.beforeCache);
    document.removeEventListener("submit", this.onSubmit);
  }
}
