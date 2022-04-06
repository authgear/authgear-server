export function setupMessageBar(): () => void {
  const disposers: Array<() => void> = [];
  const closeButtons = document.querySelectorAll("[data-close-button-target]");

  for (let i = 0; i < closeButtons.length; i++) {
    const closeButton = closeButtons[i];

    const targetID = closeButton.getAttribute("data-close-button-target");
    if (targetID == null) {
      continue;
    }

    const target = document.getElementById(targetID);
    if (target == null) {
      continue;
    }

    const onCloseButtonClick = (e: Event) => {
      e.preventDefault();
      e.stopPropagation();
      target.classList.add("hidden");
    };

    // Close the message bar before cache the page.
    // So that the cached page does not have the message bar shown.
    // See https://github.com/authgear/authgear-server/issues/1424
    const beforeCache = () => {
      if (closeButton instanceof HTMLElement) {
        closeButton.click();
      }
    };

    closeButton.addEventListener("click", onCloseButtonClick);
    document.addEventListener("turbolinks:before-cache", beforeCache);
    disposers.push(() => {
      closeButton.removeEventListener("click", onCloseButtonClick);
      document.removeEventListener("turbolinks:before-cache", beforeCache);
    });
  }

  return () => {
    for (const disposer of disposers) {
      disposer();
    }
  };
}
