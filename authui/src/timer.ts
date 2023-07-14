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
    if (!luxonDatetime.isValid) {
      return;
    }
    const displayBeforeEls = el.querySelectorAll("[data-timer-display-before]");
    const displayAfterEls = el.querySelectorAll("[data-timer-display-after]");
    const disabledBeforeEls = el.querySelectorAll(
      "[data-timer-disabled-before]"
    );

    const render = () => {
      const now = DateTime.now();
      const isBeforeDt = now < luxonDatetime;
      displayBeforeEls.forEach(
        (el) => ((el as HTMLElement).style.display = isBeforeDt ? "" : "none")
      );
      displayAfterEls.forEach(
        (el) => ((el as HTMLElement).style.display = isBeforeDt ? "none" : "")
      );
      disabledBeforeEls.forEach((el) => {
        const wasDisabled = el.getAttribute("disabled") != null;
        (el as HTMLButtonElement).disabled =
          isBeforeDt || wasDisabled ? true : false;
      });
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
