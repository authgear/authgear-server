function toggleButtonDisabled(e: Event) {
  const input = e.currentTarget as HTMLInputElement;
  const buttonID = input.getAttribute("data-account-deletion-delete-button");
  if (buttonID == null) {
    return;
  }
  const button = document.getElementById(buttonID);
  if (!(button instanceof HTMLButtonElement)) {
    return;
  }
  const value = input.value;
  button.disabled = value !== "DELETE";
}

export function setupAccountDeletion(): () => void {
  const input = document.querySelector("[data-account-deletion-delete-button]");
  if (input != null) {
    input.addEventListener("input", toggleButtonDisabled);
  }
  return () => {
    if (input != null) {
      input.removeEventListener("input", toggleButtonDisabled);
    }
  };
}
