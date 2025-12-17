import { visit, cache, session, PageSnapshot } from "@hotwired/turbo";
import { Controller } from "@hotwired/stimulus";
import axios from "axios";
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
// See https://github.com/authgear/authgear-server/issues/2333
//
// To disable the builtin form submission of Turbo,
// we call `Turbo.setFormMode("off")`.
export class TurboFormController extends Controller {
  forms: HTMLFormElement[] = [];

  // eslint-disable-next-line complexity
  async submitForm(e: Event) {
    e.preventDefault();
    const form = e.currentTarget;
    if (!(form instanceof HTMLFormElement)) {
      throw new Error("expected event.currentTarget to be a HTMLFormElement");
    }

    // Do not stop propagation so that GTM can recognize the event as Form Submission trigger.
    // e.stopPropagation();

    const formData = new FormData(form);

    const params = new URLSearchParams();
    formData.forEach((value, name) => {
      if (typeof value === "string") {
        params.set(name, value);
      } else {
        console.error("ignoring non-string value: ", name);
      }
    });
    let loadingButton: HTMLButtonElement | null = null;
    // FormData does not include any submit button's data:
    // include them manually, since we have at most one submit button per form.
    const submitButtons = form.querySelectorAll('button[type="submit"]');
    for (let i = 0; i < submitButtons.length; i++) {
      // eslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
      const button = submitButtons[i] as HTMLButtonElement;
      params.set(button.name, button.value);
      loadingButton = button;
    }
    if (form.id) {
      const el = document.querySelector(
        `button[type="submit"][form="${form.id}"]`
      );
      if (el) {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
        const button = el as HTMLButtonElement;
        params.set(button.name, button.value);
        loadingButton = button;
      }
    }
    const loadingController: LoadingController | null =
      // eslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
      this.application.getControllerForElementAndIdentifier(
        document.body,
        "loading"
      ) as LoadingController | null;
    const { onError: onLoadingError, onFinally: onLoadingFinally } =
      loadingController?.startTurboFormSubmission(loadingButton) ?? {};
    try {
      const resp = await axios(form.action, {
        method: form.method,
        headers: {
          "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
          "X-Authgear-XHR": "true",
        },
        data: params,
        onUploadProgress: progressEventHandler,
        onDownloadProgress: progressEventHandler,
      });

      if (typeof resp.data === "string") {
        // Not a json, it could be an error page.
        // Display it.
        // Ref: https://github.com/hotwired/turbo/blob/main/src/core/drive/navigator.js#L92-L107
        const snapshot = PageSnapshot.fromHTMLString(resp.data);
        await session.view.renderPage(snapshot, false, true);
        session.view.clearSnapshotCache();
        return;
      }

      const { redirect_uri, action } = resp.data;

      // Without excempting current page from cache, it will still be cached
      // right after cache.clear() is called and before the redirect is performed.
      //
      // see https://github.com/hotwired/turbo/issues/193
      cache.exemptPageFromCache();
      cache.clear();

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
      form.dispatchEvent(new CustomEvent(`turbo-form:submit-end`, {}));
    }
  }
}
