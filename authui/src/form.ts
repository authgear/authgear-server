import Turbolinks from "turbolinks";
import axios, { Method } from "axios";
import {
  disableAllButtons,
  hideProgressBar,
  showProgressBar,
  progressEventHandler,
} from "./loading";
import { handleAxiosError } from "./error";

// Handle click link to submit form
// When clicking element with `data-submit-link`, it will perform click on
// element with `data-submit-form` that contains the same value
// e.g. data-submit-link="verify-identity-resend" and
//      data-submit-form="verify-identity-resend"
export function clickLinkSubmitForm(): () => void {
  const links = document.querySelectorAll("[data-submit-link]");
  const disposers: Array<() => void> = [];
  for (let i = 0; i < links.length; i++) {
    const link = links[i];
    const formName = link.getAttribute("data-submit-link");
    const formButton = document.querySelector(
      `[data-submit-form="${formName}"]`
    );
    if (formButton instanceof HTMLElement) {
      const submitForm = (e: Event) => {
        e.preventDefault();
        formButton.click();
      };
      link.addEventListener("click", submitForm);
      disposers.push(() => {
        link.removeEventListener("click", submitForm);
      });
    }
  }
  return () => {
    for (const disposer of disposers) {
      disposer();
    }
  };
}

// Handle auto form submission
export function autoSubmitForm() {
  const e = document.querySelector('[data-auto-submit="true"]');
  if (e instanceof HTMLElement) {
    e.removeAttribute("data-auto-submit");
    e.click();
  }
}

export function xhrSubmitForm(): () => void {
  let revertDisabledButtons: { (): void } | null;

  async function submitForm(e: Event) {
    if (e.defaultPrevented) {
      return;
    }
    e.preventDefault();
    e.stopPropagation();

    const form = e.currentTarget as HTMLFormElement;
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

    revertDisabledButtons = disableAllButtons();
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
      if (revertDisabledButtons) {
        revertDisabledButtons();
        revertDisabledButtons = null;
      }
    } finally {
      hideProgressBar();
    }
  }

  const elems = document.querySelectorAll("form");
  const forms: HTMLFormElement[] = [];
  for (let i = 0; i < elems.length; i++) {
    if (elems[i].querySelector('[data-form-xhr="false"]')) {
      continue;
    }
    forms.push(elems[i] as HTMLFormElement);
  }
  for (const form of forms) {
    form.addEventListener("submit", submitForm);
  }

  // Revert disabled buttons before turbolinks cache the page
  // To avoid flickering in the UI
  const beforeCache = () => {
    if (revertDisabledButtons) {
      revertDisabledButtons();
    }
  };
  document.addEventListener("turbolinks:before-cache", beforeCache);
  return () => {
    document.removeEventListener("turbolinks:before-cache", beforeCache);
    for (const form of forms) {
      form.removeEventListener("submit", submitForm);
    }
  };
}

export function restoreForm() {
  const metaTag = document.querySelector(`meta[name="x-form-json"]`);
  if (!(metaTag instanceof HTMLMetaElement)) {
    return;
  }

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
