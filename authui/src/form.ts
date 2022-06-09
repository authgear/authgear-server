import { visit, clearCache } from "@hotwired/turbo";
import axios, { Method } from "axios";
import {
  disableAllButtons,
  hideProgressBar,
  showProgressBar,
  progressEventHandler,
} from "./loading";
import { handleAxiosError } from "./error";
import { Controller } from "@hotwired/stimulus";

export class XHRSubmitFormController extends Controller {
  revertDisabledButtons: { (): void } | null = null;
  forms: HTMLFormElement[] = [];

  // Revert disabled buttons before Turbo caches the page
  // To avoid flickering in the UI
  beforeCache = () => {
    if (this.revertDisabledButtons) {
      this.revertDisabledButtons();
    }
  };

  onSubmit = (e: Event) => {
    this.submitForm(e);
  };

  async submitForm(e: Event) {
    const form = e.currentTarget as HTMLFormElement;

    if (form.querySelector('[data-form-xhr="false"]')) {
      return;
    }

    if (e.defaultPrevented) {
      return;
    }
    e.preventDefault();
    // Do not stop propagation so that GTM can recognize the event as Form Submission trigger.
    // e.stopPropagation();

    const formData = new FormData(form);

    const params = new URLSearchParams();
    formData.forEach((value, name) => {
      params.set(name, value as string);
    });
    // FormData does not include any submit button's data:
    // include them manually, since we have at most one submit button per form.
    const submitButtons = form.querySelectorAll('button[type="submit"]');
    for (let i = 0; i < submitButtons.length; i++) {
      const button = submitButtons[i] as HTMLButtonElement;
      params.set(button.name, button.value);
    }
    if (form.id) {
      const el = document.querySelector(
        `button[type="submit"][form="${form.id}"]`
      );
      if (el) {
        const button = el as HTMLButtonElement;
        params.set(button.name, button.value);
      }
    }

    this.revertDisabledButtons = disableAllButtons();
    showProgressBar();
    try {
      const resp = await axios(form.action, {
        method: form.method as Method,
        headers: {
          "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
          "X-Authgear-XHR": "true",
        },
        data: params,
        onUploadProgress: progressEventHandler,
        onDownloadProgress: progressEventHandler,
      });

      const { redirect_uri, action } = resp.data;

      clearCache();
      switch (action) {
        case "redirect":
          // Perform full redirect.
          window.location = redirect_uri;
          break;

        case "replace":
        case "advance":
          visit(redirect_uri, { action });
          break;
      }
    } catch (e: unknown) {
      handleAxiosError(e);
      // revert is only called for error branch because
      // The success branch also loads a new page.
      // Keeping the buttons in disabled state reduce flickering in the UI.
      if (this.revertDisabledButtons) {
        this.revertDisabledButtons();
        this.revertDisabledButtons = null;
      }
    } finally {
      hideProgressBar();
    }
  }

  connect() {
    const elems = document.querySelectorAll("form");
    for (let i = 0; i < elems.length; i++) {
      if (elems[i].querySelector('[data-form-xhr="false"]')) {
        continue;
      }
      this.forms.push(elems[i] as HTMLFormElement);
    }
    for (const form of this.forms) {
      form.addEventListener("submit", this.onSubmit);
    }

    document.addEventListener("turbo:before-cache", this.beforeCache);
  }

  disconnect() {
    for (const form of this.forms) {
      form.removeEventListener("submit", this.onSubmit);
    }

    document.removeEventListener("turbo:before-cache", this.beforeCache);
  }
}

export class RestoreFormController extends Controller {
  static targets = ["metaTag"];

  declare metaTagTarget: HTMLMetaElement;

  connect() {
    const metaTag = this.metaTagTarget;

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
