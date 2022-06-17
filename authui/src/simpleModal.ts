import { Controller, ActionEvent } from "@hotwired/stimulus";

export class SimpleModalController extends Controller {
  static targets = ["modal"];
  declare modalTargets: HTMLElement[];

  changeModalVisibility(modalID: string | null, show: boolean) {
    for (let i = 0; i < this.modalTargets.length; i++) {
      const curModalID = this.modalTargets[i].getAttribute(
        "data-simple-modal-id"
      );
      if (modalID !== curModalID) {
        continue;
      }
      if (show) {
        this.modalTargets[i].classList.remove("closed");
      } else {
        this.modalTargets[i].classList.add("closed");
      }
    }
  }

  closeModal = (e: ActionEvent) => {
    e.preventDefault();
    const targetModalID = e.params.id;
    this.changeModalVisibility(targetModalID, false);
  };

  showModal = (e: ActionEvent) => {
    e.preventDefault();
    const targetModalID = e.params.id;
    this.changeModalVisibility(targetModalID, true);
  };
}
