import { Controller } from "@hotwired/stimulus";

export class DelayedAutoSubmitController extends Controller {
  declare buttonTarget: HTMLButtonElement;
  intervalReq: number | null = null;

  connect() {
    this.buttonTarget = this.element as HTMLButtonElement;
    if (!this.startCountDownWhenVisible()) {
      document.addEventListener(
        "visibilitychange",
        this.startCountDownWhenVisible
      );
    }
  }

  disconnect() {
    this.cancelAllListener();
  }

  cancelAllListener() {
    if (this.intervalReq) {
      clearInterval(this.intervalReq);
      this.intervalReq = null;
    }
    document.removeEventListener(
      "visibilitychange",
      this.startCountDownWhenVisible
    );
  }

  onClick = (e: Event) => {
    this.cancelAllListener();
  };

  setupAutoSubmit() {
    let countDownSec = Number(
      this.buttonTarget.getAttribute("data-countdown-sec")
    );
    this.intervalReq = window.setInterval(() => {
      this.buttonTarget.click();
      this.intervalReq = null;
    }, countDownSec * 1000);
  }

  startCountDownWhenVisible = () => {
    if (document.visibilityState == "visible") {
      this.setupAutoSubmit();
      // remove the listener once the counter is started
      document.removeEventListener(
        "visibilitychange",
        this.startCountDownWhenVisible
      );
      return true;
    }
    return false;
  };
}
