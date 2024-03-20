import { Controller } from "@hotwired/stimulus";

export class BodyScrollLockController extends Controller {
  scrollPosition = 0;

  lock() {
    this.scrollPosition = document.documentElement.scrollTop;
    document.body.style.overflow = "hidden";
    document.body.style.position = "fixed";
    document.body.style.top = `-${this.scrollPosition}px`;
    document.body.style.width = "100%";
    document.body.style.height = window.visualViewport?.height
      ? `${window.visualViewport.height}px`
      : "100vh";
  }

  unlock() {
    document.body.style.removeProperty("overflow");
    document.body.style.removeProperty("position");
    document.body.style.removeProperty("top");
    document.body.style.removeProperty("width");
    document.body.style.removeProperty("height");
    window.scrollTo(0, this.scrollPosition);
  }
}
