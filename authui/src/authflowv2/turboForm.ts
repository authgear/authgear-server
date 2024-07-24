import { visit, clearCache } from "@hotwired/turbo";
import { Controller } from "@hotwired/stimulus";
import axios, { Method } from "axios";
import { progressEventHandler } from "../loading";
import { LoadingController } from "./loading";
import { handleAxiosError } from "./alert-message";

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
export class TurboFormController extends Controller {
  forms: HTMLFormElement[] = [];

  async submitForm(e: Event) {
    e.preventDefault();
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
    let loadingButton: HTMLButtonElement | null = null;
    // FormData does not include any submit button's data:
    // include them manually, since we have at most one submit button per form.
    const submitButtons = form.querySelectorAll('button[type="submit"]');
    for (let i = 0; i < submitButtons.length; i++) {
      const button = submitButtons[i] as HTMLButtonElement;
      params.set(button.name, button.value);
      loadingButton = button;
    }
    if (form.id) {
      const el = document.querySelector(
        `button[type="submit"][form="${form.id}"]`
      );
      if (el) {
        const button = el as HTMLButtonElement;
        params.set(button.name, button.value);
        loadingButton = button;
      }
    }
    const loadingController: LoadingController | null =
      this.application.getControllerForElementAndIdentifier(
        document.body,
        "loading"
      ) as LoadingController | null;
    const { onError: onLoadingError, onFinally: onLoadingFinally } =
      loadingController?.startLoading(loadingButton) ?? {};
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
      onLoadingError?.();
    } finally {
      onLoadingFinally?.();
    }
  }
}
