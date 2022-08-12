import { Controller } from "@hotwired/stimulus";

// We used to disable buttons during form submission.
// It is now supported by Turbo natively.
// See https://turbo.hotwired.dev/handbook/drive#form-submissions
export class TurboformController extends Controller {
  submitEnd(e: CustomEvent) {
    const submitter: HTMLElement | undefined =
      e.detail.formSubmission.submitter;
    const formElement: HTMLFormElement = e.detail.formSubmission.formElement;
    const location: URL | undefined =
      e.detail.formSubmission.result?.fetchResponse.location;
    if (location != null) {
      const action = location.searchParams.get("x_turbo_action");
      if (action != null && action !== "") {
        if (submitter != null) {
          submitter.setAttribute("data-turbo-action", action);
        } else {
          formElement.setAttribute("data-turbo-action", action);
        }
      }
    }
  }
}

export class RestoreFormController extends Controller {
  connect() {
    const metaTag = this.element as HTMLMetaElement;

    const content = metaTag.content;
    if (content === "") {
      return;
    }

    // Clear the content to avoid restoring twice.
    metaTag.content = "";

    const formDataJSON = JSON.parse(content);

    // Find the form.
    let form: HTMLFormElement | null = null;
    const xAction = formDataJSON["x_action"];
    const elementsWithXAction = document.querySelectorAll(`[name="x_action"]`);
    for (let i = 0; i < elementsWithXAction.length; i++) {
      const elem = elementsWithXAction[i];
      if (elem instanceof HTMLButtonElement && elem.value === xAction) {
        form = elem.form;
        break;
      }
    }
    if (form == null) {
      return;
    }

    for (let i = 0; i < form.elements.length; i++) {
      const elem = form.elements[i];
      if (
        elem instanceof HTMLInputElement ||
        elem instanceof HTMLSelectElement ||
        elem instanceof HTMLTextAreaElement
      ) {
        const value = formDataJSON[elem.name];
        if (value != null) {
          elem.value = value;
        }
      }
    }
  }
}

// RetainFormFormController exposes the input listener
// to capture form on every key stroke.
// It saves the form in sessionStorage on disconnect.
// On connect, it revives the input controls by setting their value attribute.
export class RetainFormFormController extends Controller {
  static values = {
    id: String,
  };

  static targets = ["input"];

  declare idValue: string;
  declare inputTargets: HTMLInputElement[];

  retained: Record<string, string> = {};

  input(e: CustomEvent) {
    const name: string | undefined | null = (e as any).params.name;
    const value: string | undefined | null = e.detail.value;
    if (typeof name === "string" && typeof value === "string") {
      this.retained[name] = value;
    }
  }

  getSessionStorageKey(id: string): string {
    return `restore-form-form-${id}`;
  }

  connect() {
    if (this.idValue !== "") {
      const key = this.getSessionStorageKey(this.idValue);
      const value = sessionStorage.getItem(key);
      if (value == null) {
        return;
      }
      sessionStorage.removeItem(key);
      this.retained = JSON.parse(value);
      for (const input of this.inputTargets) {
        for (const name in this.retained) {
          if (input.getAttribute("data-retain-form-form-name-param") === name) {
            input.setAttribute("value", this.retained[name]);
          }
        }
      }
    }
  }

  disconnect() {
    if (this.idValue !== "") {
      const key = this.getSessionStorageKey(this.idValue);
      sessionStorage.setItem(key, JSON.stringify(this.retained));
    }
  }
}

// RetainFormInputController is intended to be installed on
// normal <input> and forward the "input" event to RetainFormFormController.
export class RetainFormInputController extends Controller {
  input(e: InputEvent) {
    if (e.currentTarget instanceof HTMLInputElement) {
      this.dispatch("input", {
        detail: {
          value: e.currentTarget.value,
        },
      });
    }
  }
}
