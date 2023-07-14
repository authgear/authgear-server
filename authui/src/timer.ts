import { Controller } from "@hotwired/stimulus";
import { DateTime } from "luxon";

export class TimerController extends Controller {
  declare intervalHandle: number | null;

  connect() {
    const el = this.element as HTMLElement;
    const rfc3339 = el.getAttribute("data-timer-until");
    if (rfc3339 == null) {
      return;
    }
    const luxonDatetime = DateTime.fromISO(rfc3339);
    const beforeEls = el.querySelectorAll("[data-timer-display-before]");
    const afterEls = el.querySelectorAll("[data-timer-display-after]");

    const render = () => {
      const now = DateTime.now();
      const isBeforeDt = now < luxonDatetime;
      beforeEls.forEach(
        (el) => ((el as HTMLElement).style.display = isBeforeDt ? "" : "none")
      );
      afterEls.forEach(
        (el) => ((el as HTMLElement).style.display = isBeforeDt ? "none" : "")
      );
    };
    render();

    this.intervalHandle = window.setInterval(render, 100);
  }

  disconnect() {
    if (this.intervalHandle != null) {
      window.clearInterval(this.intervalHandle);
      this.intervalHandle = null;
    }
  }
}
