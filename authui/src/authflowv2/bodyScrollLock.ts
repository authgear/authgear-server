import { Controller } from "@hotwired/stimulus";

export class BodyScrollLockController extends Controller {
  scrollPosition = 0;
  styleElement: HTMLStyleElement | null = null;

  get style() {
    const height = window.visualViewport
      ? `${window.visualViewport.height}px`
      : "100vh";

    return `
      body {
        position: fixed;
        top: -${this.scrollPosition}px;
        left: 0;
        right: 0;
        width: 100%;
        height: ${height};
      }
    `;
  }

  lock() {
    this.scrollPosition = document.documentElement.scrollTop;

    this.styleElement = document.createElement("style");
    document.head.appendChild(this.styleElement);
    this.styleElement.sheet?.insertRule(this.style, 0);
  }

  unlock() {
    if (this.styleElement) {
      document.head.removeChild(this.styleElement);
      this.styleElement = null;
    }

    window.scrollTo(0, this.scrollPosition);
  }
}
