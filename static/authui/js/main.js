window.addEventListener("load", function() {
  // global variables
  // They must be reset to their initial values when back-forward cache kicks in.
  var FORM_SUBMITTED;

  function initializeGlobals() {
    FORM_SUBMITTED = false;
  }
  initializeGlobals();

  window.addEventListener("pageshow", function(e) {
    if (e.persisted) {
      initializeGlobals();
    }
  });

  function togglePasswordVisibility() {
    var pwd = document.querySelector("#password");
    if (pwd == null) {
      return;
    }
    if (pwd.type === "password") {
      pwd.type = "text";
    } else {
      pwd.type = "password";
    }
    var els = document.querySelectorAll(".password-visibility-btn");
    for (var i = 0; i < els.length; i++) {
      var el = els[i];
      if (el.classList.contains("show-password")) {
        if (pwd.type === "text") {
          el.style.display = "none";
        } else {
          el.style.display = "block";
        }
      }
      if (el.classList.contains("hide-password")) {
        if (pwd.type === "password") {
          el.style.display = "none";
        } else {
          el.style.display = "block";
        }
      }
    }
  }

  function attachPasswordVisibilityClick() {
    var els = document.querySelectorAll(".password-visibility-btn");
    for (var i = 0; i < els.length; i++) {
      var el = els[i];
      el.addEventListener("click", function(e) {
        e.preventDefault();
        e.stopPropagation();
        togglePasswordVisibility();
      });
    }
  }

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
      el.classList.add("passed");
    }
  }

  function checkPasswordUppercase(value, el) {
    if (el == null) {
      return;
    }
    if (/[A-Z]/.test(value)) {
      el.classList.add("passed");
    }
  }

  function checkPasswordLowercase(value, el) {
    if (el == null) {
      return;
    }
    if (/[a-z]/.test(value)) {
      el.classList.add("passed");
    }
  }

  function checkPasswordDigit(value, el) {
    if (el == null) {
      return;
    }
    if (/[0-9]/.test(value)) {
      el.classList.add("passed");
    }
  }

  function checkPasswordSymbol(value, el) {
    if (el == null) {
      return;
    }
    if (/[^a-zA-Z0-9]/.test(value)) {
      el.classList.add("passed");
    }
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
        els[i].classList.remove("violated", "passed");
      }
      checkPasswordLength(value, document.querySelector(".password-policy.length"));
      checkPasswordUppercase(value, document.querySelector(".password-policy.uppercase"));
      checkPasswordLowercase(value, document.querySelector(".password-policy.lowercase"));
      checkPasswordDigit(value, document.querySelector(".password-policy.digit"));
      checkPasswordSymbol(value, document.querySelector(".password-policy.symbol"));
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
      if (timeElapsed > cooldown) {
        el.disabled = false;
      } else {
        el.disabled = true;
        displaySeconds = Math.round((cooldown - timeElapsed) / 1000);
      }

      if (displaySeconds === 0) {
        el.textContent = label;
      } else {
        el.textContent = labelUnit.replace("%d", String(displaySeconds));
        requestAnimationFrame(tick);
      }
    }

    requestAnimationFrame(tick);
  }

  // Disable all form submission if any form has been submitted once.
  function attachFormSubmit() {
    var els = document.querySelectorAll("form");
    for (var i = 0; i < els.length; ++i) {
      var form = els[i];
      form.addEventListener("submit", function(e) {
        if (!FORM_SUBMITTED) {
          FORM_SUBMITTED = true;
        } else {
          e.preventDefault();
          e.stopPropagation();
        }
      });
    }
  }

  attachPasswordVisibilityClick();
  attachBackButtonClick();
  attachPasswordPolicyCheck();
  attachResendButtonBehavior();
  attachFormSubmit();
});
