// usage: adding follow data attribute in the button element
// - data-modal="confirmation"
// - data-modal-title="{TITLE_TEXT}"
// - data-modal-body="{BODY_TEXT}"
// - data-modal-action-label="{ACTION_LABEL_TEXT}"
// - data-modal-cancel-label="{CANCEL_LABEL_TEXT}"
export function setupModal(): () => void {
  const modal = document.querySelector('[data-modal-ele="true"]');
  if (!(modal instanceof HTMLElement)) {
    // modal template not found
    return () => {};
  }

  const modalTitleEle = modal.querySelector(
    '[data-modal-title-ele="true"]'
  ) as HTMLElement;
  const modalBodyEle = modal.querySelector(
    '[data-modal-body-ele="true"]'
  ) as HTMLElement;
  const modalActionBtnEle = modal.querySelector(
    '[data-modal-action-btn-ele="true"]'
  ) as HTMLElement;
  const modalCancelBtnEle = modal.querySelector(
    '[data-modal-cancel-btn-ele="true"]'
  ) as HTMLElement;
  const modalOverlayEle = modal.querySelector(
    '[data-modal-overlay-ele="true"]'
  ) as HTMLElement;

  const buttons = document.querySelectorAll('[data-modal="confirmation"]');
  const disposers: Array<() => void> = [];
  var confirmed = false;

  for (let i = 0; i < buttons.length; i++) {
    const button = buttons[i] as HTMLElement;

    const closeModal = () => {
      confirmed = false;
      disposeModal();
      modal.classList.add("closed");
    };

    const onClickModalAction = (e: Event) => {
      confirmed = true;
      button.click();
    };

    const onClickModalCancel = (e: Event) => {
      closeModal();
    };

    const disposeModal = () => {
      modalActionBtnEle.removeEventListener("click", onClickModalAction);
      modalCancelBtnEle.removeEventListener("click", onClickModalCancel);
      modalOverlayEle.removeEventListener("click", onClickModalCancel);
    };

    const confirmFormSubmit = (e: Event) => {
      if (confirmed) {
        // close the modal and perform the default behaviour
        closeModal();
        return;
      }
      e.preventDefault();
      modalTitleEle.innerText = button.dataset["modalTitle"] || "";
      modalBodyEle.innerText = button.dataset["modalBody"] || "";
      modalActionBtnEle.innerText = button.dataset["modalActionLabel"] || "";
      modalCancelBtnEle.innerText = button.dataset["modalCancelLabel"] || "";

      modalActionBtnEle.addEventListener("click", onClickModalAction);
      modalCancelBtnEle.addEventListener("click", onClickModalCancel);
      modalOverlayEle.addEventListener("click", onClickModalCancel);

      modal.classList.remove("closed");
    };

    button.addEventListener("click", confirmFormSubmit);
    disposers.push(() => {
      button.removeEventListener("click", confirmFormSubmit);
      disposeModal();
    });
  }

  return () => {
    for (const disposer of disposers) {
      disposer();
    }
  };
}
