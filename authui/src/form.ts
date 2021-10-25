import Turbolinks from "turbolinks";

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
  let isSubmitting = false;
  function submitForm(e: Event) {
    e.preventDefault();
    e.stopPropagation();
    if (isSubmitting) {
      return;
    }
    isSubmitting = true;

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

    fetch(form.action, {
      method: form.method,
      headers: {
        "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
        "X-Authgear-XHR": "true",
      },
      credentials: "same-origin",
      body: params,
    })
      .then((resp) => {
        if (resp.status < 200 || resp.status >= 300) {
          isSubmitting = false;
          setServerError();
          return;
        }
        return resp.json().then(({ redirect_uri, action }) => {
          isSubmitting = false;

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
        });
      })
      .catch(() => {
        isSubmitting = false;
        setNetworkError();
      });
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

  return () => {
    for (const form of forms) {
      form.removeEventListener("submit", submitForm);
    }
  };
}

function setServerError() {
  const errorBar = document.querySelector(".errors-messages-bar");
  if (errorBar) {
    const list = errorBar.querySelector(".error-txt");
    var li = document.createElement("li");
    li.appendChild(
      document.createTextNode(errorBar.getAttribute("data-server-error") || "")
    );
    if (list) list.innerHTML = li.outerHTML;
    errorBar.classList.add("flex");
    errorBar.classList.remove("hidden");
  }
}

function setNetworkError() {
  const errorBar = document.querySelector(".errors-messages-bar");
  if (errorBar) {
    const list = errorBar.querySelector(".error-txt");
    var li = document.createElement("li");
    li.appendChild(
      document.createTextNode(errorBar.getAttribute("data-network-error") || "")
    );
    if (list) list.innerHTML = li.outerHTML;
    errorBar.classList.add("flex");
    errorBar.classList.remove("hidden");
  }
}
