export function setupPasswordVisibilityToggle(): () => void {
  const passwordInputs = document.querySelectorAll(
    "[data-show-password-button]"
  );

  const disposers: Array<() => void> = [];

  for (let i = 0; i < passwordInputs.length; i++) {
    const passwordInput = passwordInputs[i] as HTMLInputElement;

    const showPasswordButtonID = passwordInput.getAttribute(
      "data-show-password-button"
    );
    const hidePasswordButtonID = passwordInput.getAttribute(
      "data-hide-password-button"
    );
    if (showPasswordButtonID == null || hidePasswordButtonID == null) {
      continue;
    }

    const showPasswordButton = document.getElementById(showPasswordButtonID);
    const hidePasswordButton = document.getElementById(hidePasswordButtonID);
    if (showPasswordButton == null || hidePasswordButton == null) {
      continue;
    }

    const togglePasswordVisibility = (e: Event) => {
      e.preventDefault();
      e.stopPropagation();

      if (hidePasswordButton.classList.contains("hidden")) {
        passwordInput.type = "text";
        showPasswordButton.classList.add("hidden");
        hidePasswordButton.classList.remove("hidden");
      } else {
        passwordInput.type = "password";
        showPasswordButton.classList.remove("hidden");
        hidePasswordButton.classList.add("hidden");
      }
    };

    showPasswordButton.addEventListener("click", togglePasswordVisibility);
    hidePasswordButton.addEventListener("click", togglePasswordVisibility);
    disposers.push(() => {
      showPasswordButton.removeEventListener("click", togglePasswordVisibility);
      hidePasswordButton.removeEventListener("click", togglePasswordVisibility);
    });
  }

  return () => {
    for (const disposer of disposers) {
      disposer();
    }
  };
}
