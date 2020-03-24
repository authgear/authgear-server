window.addEventListener("load", function() {
  function setupPasswordVisibilityButton(el) {
    if (el == null) {
      return;
    }
    var pwd = document.querySelector("#password");
    if (pwd == null) {
      return;
    }
    if (pwd.type === "password") {
      el.classList.remove("hide-password");
      el.classList.add("show-password");
      el.textContent = "Show Password";
    } else if (pwd.type === "text") {
      el.classList.remove("show-password");
      el.classList.add("hide-password");
      el.textContent = "Hide Password";
    }
  }

  function togglePasswordVisibility(el) {
    if (el == null) {
      return;
    }
    var pwd = document.querySelector("#password");
    if (pwd == null) {
      return;
    }
    if (pwd.type === "password") {
      pwd.type = "text";
    } else {
      pwd.type = "password";
    }
    setupPasswordVisibilityButton(el);
  }

  function attachPasswordVisibilityClick() {
    var el = document.querySelector(".btn.toggle-password-visibility");
    if (el != null) {
      setupPasswordVisibilityButton(el);
      el.addEventListener("click", function(e) {
        e.preventDefault();
        e.stopPropagation();
        togglePasswordVisibility(el);
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

  attachPasswordVisibilityClick();
  attachBackButtonClick();
  attachPasswordPolicyCheck();
});
