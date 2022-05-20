import { Controller } from "@hotwired/stimulus";

// usage: adding follow data attribute in the button element
// - data-modal-target="submitButton"
// - data-action="click->modal#confirmFormSubmit"
// - data-modal-title="{TITLE_TEXT}"
// - data-modal-body="{BODY_TEXT}"
// - data-modal-action-label="{ACTION_LABEL_TEXT}"
// - data-modal-cancel-label="{CANCEL_LABEL_TEXT}"
export class ModalController extends Controller {
  static targets = [
    "modalEle",
    "modalTitleEle",
    "modalBodyEle",
    "modalActionBtnEle",
    "modalCancelBtnEle",
    "modalOverlayEle",
  ];

  declare modalEleTarget: HTMLElement;
  declare modalTitleEleTarget: HTMLElement;
  declare modalBodyEleTarget: HTMLElement;
  declare modalActionBtnEleTarget: HTMLElement;
  declare modalCancelBtnEleTarget: HTMLElement;
  declare modalOverlayEleTarget: HTMLElement;

  submitButton: HTMLButtonElement | null = null;
  confirmed: boolean = false;

  closeModal = () => {
    this.confirmed = false;
    this.modalEleTarget.classList.add("closed");
  };

  onClickModalAction(e: Event) {
    this.confirmed = true;
    this.submitButton?.click();
  }

  onClickModalCancel(e: Event) {
    this.closeModal();
  }

  confirmFormSubmit(e: Event) {
    if (this.confirmed) {
      // close the modal and perform the default behaviour
      this.closeModal();
      return;
    }
    e.preventDefault();

    this.submitButton = e.currentTarget as HTMLButtonElement;

    const button = this.submitButton;
    const modalTitleEle = this.modalTitleEleTarget;
    const modalBodyEle = this.modalBodyEleTarget;
    const modalActionBtnEle = this.modalActionBtnEleTarget;
    const modalCancelBtnEle = this.modalCancelBtnEleTarget;

    modalTitleEle.innerText = button.dataset["modalTitle"] || "";
    modalBodyEle.innerText = button.dataset["modalBody"] || "";
    modalActionBtnEle.innerText = button.dataset["modalActionLabel"] || "";
    modalCancelBtnEle.innerText = button.dataset["modalCancelLabel"] || "";

    this.modalEleTarget.classList.remove("closed");
  }
}
