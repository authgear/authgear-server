import { Controller } from "@hotwired/stimulus";

export class BackController extends Controller {
  connect(): void {
    this.element.classList.remove("hidden");
    this.element.addEventListener("click", this.back);
  }

  disconnect(): void {
    this.element.removeEventListener("click", this.back);
  }

  back = () => {
    window.history.back();
  };
}
