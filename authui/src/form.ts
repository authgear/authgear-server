import { visit, clearCache } from "@hotwired/turbo";
import { Controller } from "@hotwired/stimulus";
import axios, { Method } from "axios";
import {
  disableAllButtons,
  hideProgressBar,
  showProgressBar,
  progressEventHandler,
} from "./loading";
import { handleAxiosError } from "./messageBar";

// Turbo has builtin support for form submission.
// We once migrated to use it.
// However, redirect to external is broken because
// the redirect is made with fetch, which is subject to CORS.
// A typical problem is the support for the post login redirect URI.
// The post login redirect URI is usually an external link that
// the origin of the link does not list our origin as allowed origin.
// If we rely on Turbo to handle form submission,
// the post login redirect URI will be broken.
// Therefore, we reverted to handle form submission ourselves.
// To disable the builtin form submission of Turbo,
// I studied its source code and discovered that
// Turbo checked in the bubble event listener to see if
// the event is prevented.
// So I added a capture event listener to call preventDefault()
// to stop Turbo from submitting forms.
//
// See https://github.com/authgear/authgear-server/issues/2333
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

  onSubmitCapture = (e: Event) => {
    e.preventDefault();
  };

  onSubmit = (e: Event) => {
    this.submitForm(e);
  };

  async submitForm(e: Event) {
    const form = e.currentTarget as HTMLFormElement;

    if (form.querySelector('[data-turbo="false"]')) {
      return;
    }

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
          // We assume the URL returned by the server have query preserved.
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
      if (elems[i].querySelector('[data-turbo="false"]')) {
        continue;
      }
      this.forms.push(elems[i] as HTMLFormElement);
    }
    for (const form of this.forms) {
      form.addEventListener("submit", this.onSubmitCapture, true);
      form.addEventListener("submit", this.onSubmit);
    }

    document.addEventListener("turbo:before-cache", this.beforeCache);
  }

  disconnect() {
    for (const form of this.forms) {
      form.removeEventListener("submit", this.onSubmitCapture, true);
      form.removeEventListener("submit", this.onSubmit);
    }

    document.removeEventListener("turbo:before-cache", this.beforeCache);
  }
}

// NOTE: As turbo would first disconnect and connect head and then disconnect and connect body, if we put this controller in head and modify field value, it will then be covered.
export class RestoreFormController extends Controller {
  static values = { json: String };

  declare jsonValue: string;

  connect() {
    const content = this.jsonValue;
    if (content === "") {
      return;
    }

    // Clear the content to avoid restoring twice.
    this.jsonValue = "";

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
    if (form.getAttribute("data-restore-form") === "false") {
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
            // NOTE(tung): Setting value attribute cause bfcache of chrome failing to restore
            // input.setAttribute("value", this.retained[name]);
            input.value = this.retained[name];
          }
        }
      }
    }
  }

  disconnect() {
    // Before disconnect, collect all values once.
    this.collectAllValues();
    if (this.idValue !== "") {
      const key = this.getSessionStorageKey(this.idValue);
      sessionStorage.setItem(key, JSON.stringify(this.retained));
    }
  }

  private collectAllValues() {
    for (const input of this.inputTargets) {
      const name = input.getAttribute("data-retain-form-form-name-param");
      const value = input.value;
      if (typeof name === "string" && typeof value === "string") {
        this.retained[name] = value;
      }
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
