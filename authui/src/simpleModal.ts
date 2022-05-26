import { Controller } from "@hotwired/stimulus";

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

  closeModal = (e: Event) => {
    e.preventDefault();
    const ele = e.currentTarget as HTMLElement;
    const targetModalID = ele.getAttribute("data-simple-modal-id");
    this.changeModalVisibility(targetModalID, false);
  };

  showModal = (e: Event) => {
    e.preventDefault();
    const ele = e.currentTarget as HTMLElement;
    const targetModalID = ele.getAttribute("data-simple-modal-id");
    this.changeModalVisibility(targetModalID, true);
  };
}
