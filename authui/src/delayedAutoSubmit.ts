import { Controller } from "@hotwired/stimulus";

export class DelayedAutoSubmitController extends Controller {
  static targets = ["button"];

  declare buttonTarget: HTMLButtonElement;
  countDownSec: number = 0;

  animationReq: number | null = null;
  intervalReq: number | null = null;
  visibilityChangeListener: (() => void) | null = null;

  connect() {
    this.countDownSec = Number(
      this.buttonTarget.getAttribute("data-countdown-sec")
    );
    if (!this.startCountDownWhenVisible()) {
      this.visibilityChangeListener = this.startCountDownWhenVisible.bind(this);
      document.addEventListener(
        "visibilitychange",
        this.visibilityChangeListener
      );
    }
  }

  disconnect() {
    if (this.animationReq) {
      cancelAnimationFrame(this.animationReq);
    }
    if (this.intervalReq) {
      cancelAnimationFrame;
    }
    if (this.visibilityChangeListener) {
      document.removeEventListener(
        "visibilitychange",
        this.visibilityChangeListener
      );
    }
  }

  startCounter() {
    const label = this.buttonTarget.getAttribute("data-label");
    const labelUnit = this.buttonTarget.getAttribute("data-label-unit");
    const endAt = new Date().getTime() + this.countDownSec * 1000;
    const that = this;

    function count() {
      const remainingTime = endAt - new Date().getTime();
      let displaySeconds = Math.round(remainingTime / 1000);
      if (displaySeconds < 0) {
        that.buttonTarget.disabled = true;
        that.buttonTarget.textContent = label;
        that.animationReq = null;
      } else {
        that.buttonTarget.textContent =
          labelUnit && labelUnit.replace("%d", String(displaySeconds));
      }
      that.animationReq = requestAnimationFrame(count);
    }
    this.animationReq = requestAnimationFrame(count);
  }

  setupAutoSubmit() {
    this.intervalReq = setInterval(() => {
      this.buttonTarget.click();
    }, this.countDownSec * 1000);
  }

  startCountDownWhenVisible() {
    if (document.visibilityState == "visible") {
      this.startCounter();
      this.setupAutoSubmit();
      return true;
    }
    return false;
  }
}
