import Turbolinks from "turbolinks";
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

  // Revert disabled buttons before turbolinks cache the page
  // To avoid flickering in the UI
  beforeCache = () => {
    if (this.revertDisabledButtons) {
      this.revertDisabledButtons();
    }
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

      Turbolinks.clearCache();
      switch (action) {
        case "redirect":
          // Perform full redirect.
          window.location = redirect_uri;
          break;

        case "replace":
        case "advance":
          Turbolinks.visit(redirect_uri, { action });
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
      form.addEventListener("submit", this.submitForm);
    }

    document.addEventListener("turbolinks:before-cache", this.beforeCache);
  }

  disconnect() {
    for (const form of this.forms) {
      form.removeEventListener("submit", this.submitForm);
    }

    document.removeEventListener("turbolinks:before-cache", this.beforeCache);
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
