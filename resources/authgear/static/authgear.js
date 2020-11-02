// global variables
// They must be reset to their initial values when back-forward cache kicks in.
var FORM_SUBMITTED;

function initializeGlobals() {
  FORM_SUBMITTED = false;
}

window.addEventListener("pageshow", function(e) {
  if (e.persisted) {
    initializeGlobals();
  }
});

function attachHistoryListener() {
  window.addEventListener("popstate", function(e) {
    var meta = document.querySelector('meta[name="x-authgear-request-url"]');
    if (meta == null) {
      return;
    }

    var currentURL = new URL(window.location.href);
    // meta.content is without protocol, host, port.
    // URL constructor is not smart enough to parse it.
    // Therefore we need to tell them the base URL.
    var pageURL = new URL(meta.content, currentURL);

    if (currentURL.pathname !== pageURL.pathname) {
      window.location.reload();
    }
  });
}
attachHistoryListener();

function setupPage() {
  initializeGlobals();

  function attachBackButtonClick() {
    var els = document.querySelectorAll(".btn.back-btn");
    for (var i = 0; i < els.length; i++) {
      var el = els[i];
      el.addEventListener("click", function(e) {
        e.preventDefault();
        e.stopPropagation();
        window.history.back();
      });
    }
  }

  function checkPasswordLength(value, el) {
    if (el == null) {
      return;
    }
    var minLength = parseInt(el.getAttribute("data-min-length"), 10);
    // .length is number of UTF-16 code units,
    // while the server is counting number of UTF-8 code units.
    if (value.length >= minLength) {
      el.classList.add("good-txt");
    }
  }

  function checkPasswordUppercase(value, el) {
    if (el == null) {
      return;
    }
    if (/[A-Z]/.test(value)) {
      el.classList.add("good-txt");
    }
  }

  function checkPasswordLowercase(value, el) {
    if (el == null) {
      return;
    }
    if (/[a-z]/.test(value)) {
      el.classList.add("good-txt");
    }
  }

  function checkPasswordDigit(value, el) {
    if (el == null) {
      return;
    }
    if (/[0-9]/.test(value)) {
      el.classList.add("good-txt");
    }
  }

  function checkPasswordSymbol(value, el) {
    if (el == null) {
      return;
    }
    if (/[^a-zA-Z0-9]/.test(value)) {
      el.classList.add("good-txt");
    }
  }

  function checkPasswordStrength(value) {
    var meter = document.querySelector("#password-strength-meter");
    var desc = document.querySelector("#password-strength-meter-description");
    if (meter == null || desc == null) {
      return;
    }

    meter.value = 0;
    desc.textContent = "";

    if (value === "") {
      return;
    }

    var result = zxcvbn(value);
    var score = Math.min(5, Math.max(1, result.score + 1));
    meter.value = score;
    desc.textContent = desc.getAttribute("data-desc-" + score);
  }

  function attachPasswordPolicyCheck() {
    var el = document.querySelector("[data-password-policy-password]");
    if (el == null ) {
      return;
    }
    el.addEventListener("input", function(e) {
      var value = e.currentTarget.value;
      var els = document.querySelectorAll(".password-policy");
      for (var i = 0; i < els.length; ++i) {
        els[i].classList.remove("error-txt", "good-txt");
      }
      checkPasswordLength(value, document.querySelector(".password-policy.length"));
      checkPasswordUppercase(value, document.querySelector(".password-policy.uppercase"));
      checkPasswordLowercase(value, document.querySelector(".password-policy.lowercase"));
      checkPasswordDigit(value, document.querySelector(".password-policy.digit"));
      checkPasswordSymbol(value, document.querySelector(".password-policy.symbol"));
      checkPasswordStrength(value);
    });
  }

  function attachResendButtonBehavior() {
    var el = document.querySelector("#resend-button");
    if (el == null) {
      return;
    }


    var scheduledAt = new Date();
    var cooldown = parseInt(el.getAttribute("data-cooldown"), 10) * 1000;
    var label = el.getAttribute("data-label");
    var labelUnit = el.getAttribute("data-label-unit");

    function tick() {
      var now = new Date();
      var timeElapsed = now - scheduledAt;

      var displaySeconds = 0;
      if (timeElapsed <= cooldown) {
        displaySeconds = Math.round((cooldown - timeElapsed) / 1000);
      }

      if (displaySeconds === 0) {
        el.disabled = false;
        el.textContent = label;
      } else {
        el.disabled = true;
        el.textContent = labelUnit.replace("%d", String(displaySeconds));
        requestAnimationFrame(tick);
      }
    }

    requestAnimationFrame(tick);
  }

  // Disable all form submission if any form has been submitted once.
  function attachFormSubmitOnceOnly() {
    var els = document.querySelectorAll("form");
    for (var i = 0; i < els.length; ++i) {
      var form = els[i];

      // Allow submitting form natively multiple times
      var shouldIgnored = false;
      for (var j = 0; j < form.elements.length; ++j) {
        var field = form.elements[j];

        if (field.getAttribute("data-form-xhr") === "false") {
          shouldIgnored = true;
          break;
        }
      }
      if (shouldIgnored) {
        continue;
      }

      form.addEventListener("submit", function(e) {
        if (!FORM_SUBMITTED) {
          FORM_SUBMITTED = true;
        } else {
          e.preventDefault();
          e.stopPropagation();
          e.stopImmediatePropagation();
        }
      });
    }
  }


  // It is important that this is not an arrow function.
  // This function relies on `this` being the XHR object.
  function handleXHRFormSubmission(e) {
    var currentURLString = window.location.href;
    var responseURLString = this.responseURL;
    var responseHTML = this.responseText;

    var currentURL = new URL(currentURLString);
    var responseURL = new URL(responseURLString);

    // We are still within our application.
    if (currentURL.protocol === responseURL.protocol && currentURL.origin === responseURL.origin) {
      // Same path. Currently we assume this is form submission error.
      if (currentURL.pathname === responseURL.pathname) {
        history.replaceState({}, "", responseURLString);
      } else {
        history.pushState({}, "", responseURLString);
      }

      // Replace current document with newly loaded document.
      var parser = new DOMParser();
      var newDocument = parser.parseFromString(responseHTML, "text/html");
      document.body.replaceWith(newDocument.body);
      // Manually update content of head, to avoid flickering caused by reloading CSS.
      document.head.title = newDocument.head.title;
      var metaElements = document.querySelectorAll("meta");
      for (var i = 0; i < metaElements.length; i++) {
        metaElements[i].remove();
      }
      metaElements = newDocument.querySelectorAll("meta");
      for (var i = 0; i < metaElements.length; i++) {
        document.head.appendChild(metaElements[i]);
      }
      setupPage();

    } else {
      // Otherwise redirect natively.
      window.location.href = responseURLString;
    }
  }

  // Use XHR to submit form.
  // If we rely on the browser to submit the form for us,
  // error submission will add an entry to the history stack,
  // causing back button fail to work intuitively.
  //
  // Therefore, when JavaScript is available,
  // we use XHR to submit the form.
  // XHR follows redirect automatically
  // and .responseURL is GET URL we need to visit to retrieve the submission result.
  // If window.location.href is assigned the same value, no extra entry is added to the history stack.
  function attachFormSubmitXHR() {
    var els = document.querySelectorAll("form");
    for (var i = 0; i < els.length; ++i) {
      els[i].addEventListener("submit", function(e) {

        var shouldIgnored = false;

        var form = e.currentTarget;
        // e.submitter is not supported by Safari
        // therefore we must not have multiple submit buttons per form.

        // https://html.spec.whatwg.org/multipage/form-control-infrastructure.html#constructing-form-data-set
        var entryList = [];
        for (var j = 0; j < form.elements.length; ++j) {
          var field = form.elements[j];

          if (field.getAttribute("data-form-xhr") === "false") {
            shouldIgnored = true;
          }

          // Step 5.1 Point 1 is ignored because we do not use datalist.

          // Step 5.1 Point 2
          if (field.disabled) {
            continue;
          }

          // Step 5.1 Point 3
          // if (field instanceof HTMLButtonElement && field !== submitter) {
          //   continue;
          // }
          // Step 5.1 Point 3
          // if (field instanceof HTMLInputElement && field.type === "submit" && field !== submitter) {
          //   continue;
          // }

          // Step 5.1 Point 4
          if (field instanceof HTMLInputElement && field.type === "checkbox" && !field.checked) {
            continue;
          }

          // Step 5.1 Point 5
          if (field instanceof HTMLInputElement && field.type === "radio" && !field.checked) {
            continue;
          }

          // Step 5.1 Point 6; It deviates from the spec because we do not use <object>.
          if (field instanceof HTMLObjectElement) {
            continue;
          }

          // Step 5.2; It deviates from the spec becaues we do not use <input type="image">.
          if (field instanceof HTMLInputElement && field.type === "image") {
            continue;
          }

          // Step 5.3 is ignored because we do not use form-associated custom element.

          // Step 5.4
          if (field.name === "" || field.name == null) {
            continue;
          }

          // Step 5.5
          var name = field.name;
          var value = field.value;

          // Step 5.6 <select> is omitted for now.

          // Step 5.7
          if (field instanceof HTMLInputElement && (field.type === "checkbox" || field.type === "radio")) {
            if (field.value == null || field.value === "") {
              value = "on";
            }
          }

          // Step 5.8 is ignored because we do not use file upload.
          // Step 5.9 is ignored because we do not use <object>.
          // Step 5.10 is ignored because we do ot use <input type="hidden" name="_charset_">.
          // Step 5.11 <textarea> is omitted for now.

          // Step 5.12
          entryList.push([name, value]);

          // Step 5.13 is ignored because we do nto use dirname.
        }

        // Ignore any form containing elements with "data-form-xhr"
        // Such forms will redirect to external location
        // so CORS will kick in and XHR does not work.
        if (shouldIgnored) {
          return;
        }

        e.preventDefault();
        e.stopPropagation();

        var body = new URLSearchParams();
        for (var i = 0; i < entryList.length; ++i) {
          var entry = entryList[i];
          body.append(entry[0], entry[1]);
        }

        var xhr = new XMLHttpRequest();
        xhr.withCredentials = true;
        xhr.onload = handleXHRFormSubmission;
        xhr.open(form.method, form.action, true);
        // Safari does not support xhr.send(URLSearchParams)
        // so we have to manually set content-type
        // and serialize URLSearchParams to string.
        xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8");
        xhr.send(body.toString());
      });
    }
  }

  function attachShowAndHidePasswordButtonClickToWrapper(wrapper) {
    var input = wrapper.querySelector(".input");
    var showPasswordButton = wrapper.querySelector(".show-password-button");
    var hidePasswordButton = wrapper.querySelector(".hide-password-button");
    if (!input || !showPasswordButton || !hidePasswordButton) {
      return;
    }

    if (wrapper.classList.contains("show-password")) {
      input.type = "text";
    } else {
      input.type = "password";
    }

    showPasswordButton.addEventListener("click", function(e) {
      wrapper.classList.add("show-password");
      input.type = "text";
    });

    hidePasswordButton.addEventListener("click", function(e) {
      wrapper.classList.remove("show-password");
      input.type = "password";
    });
  }

  function attachShowAndHidePasswordButtonClick() {
    var wrappers = document.querySelectorAll(".password-input-wrapper");
    for (var i = 0; i < wrappers.length; ++i) {
      var wrapper = wrappers[i];
      attachShowAndHidePasswordButtonClickToWrapper(wrapper);
    }
  }

  attachBackButtonClick();
  attachPasswordPolicyCheck();
  attachResendButtonBehavior();
  attachFormSubmitOnceOnly();
  attachFormSubmitXHR();
  attachShowAndHidePasswordButtonClick();
}
window.addEventListener("load", setupPage);
