import { init } from "./core";
import "./authgear.css";

init();

window.api.onLoad(() => {
  document.body.classList.add("js");
});

// Handle back button.

function back(e: Event) {
  e.preventDefault();
  e.stopPropagation();
  window.history.back();
}

window.api.onLoad(() => {
  const elems = document.querySelectorAll(".back-btn");
  for (let i = 0; i < elems.length; i++) {
    elems[i].addEventListener("click", back);
  }

  return () => {
    for (let i = 0; i < elems.length; i++) {
      elems[i].removeEventListener("click", back);
    }
  };
});

// Handle password visibility toggle.

window.api.onLoad(() => {
  const wrappers = document.querySelectorAll(".password-input-wrapper");
  const disposers: Array<() => void> = [];
  for (let i = 0; i < wrappers.length; i++) {
    const wrapper = wrappers[i];
    const input = wrapper.querySelector(".input") as HTMLInputElement;
    const showPasswordButton = wrapper.querySelector(".show-password-button");
    const hidePasswordButton = wrapper.querySelector(".hide-password-button");
    if (!input || !showPasswordButton || !hidePasswordButton) {
      return;
    }

    if (wrapper.classList.contains("show-password")) {
      input.type = "text";
    } else {
      input.type = "password";
    }

    const togglePasswordPolicy = (e: Event) => {
      e.preventDefault();
      e.stopPropagation();
      wrapper.classList.toggle("show-password");
      if (wrapper.classList.contains("show-password")) {
        input.type = "text";
      } else {
        input.type = "password";
      }
    };

    showPasswordButton.addEventListener("click", togglePasswordPolicy);
    hidePasswordButton.addEventListener("click", togglePasswordPolicy);
    disposers.push(() => {
      showPasswordButton.removeEventListener("click", togglePasswordPolicy);
      hidePasswordButton.removeEventListener("click", togglePasswordPolicy);
    });
  }

  return () => {
    for (const disposer of disposers) {
      disposer();
    }
  };
});

// Handle resend button.

window.api.onLoad(() => {
  const el = document.querySelector("#resend-button") as HTMLButtonElement;
  if (el == null) {
    return;
  }

  const scheduledAt = new Date();
  const cooldown = Number(el.getAttribute("data-cooldown")) * 1000;
  const label = el.getAttribute("data-label");
  const labelUnit = el.getAttribute("data-label-unit")!;
  let animHandle: number | null = null;

  function tick() {
    const now = new Date();
    const timeElapsed = now.getTime() - scheduledAt.getTime();

    let displaySeconds = 0;
    if (timeElapsed <= cooldown) {
      displaySeconds = Math.round((cooldown - timeElapsed) / 1000);
    }

    if (displaySeconds === 0) {
      el.disabled = false;
      el.textContent = label;
      animHandle = null;
    } else {
      el.disabled = true;
      el.textContent = labelUnit.replace("%d", String(displaySeconds));
      animHandle = requestAnimationFrame(tick);
    }
  }

  animHandle = requestAnimationFrame(tick);

  return () => {
    if (animHandle != null) {
      cancelAnimationFrame(animHandle);
    }
  };
});
