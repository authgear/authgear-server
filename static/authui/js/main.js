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

  attachPasswordVisibilityClick();
});
