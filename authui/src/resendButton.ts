import { Controller } from "@hotwired/stimulus";

export class ResendButtonController extends Controller {
  static targets = ["button"];

  declare buttonTarget: HTMLButtonElement;
  declare animHandle: number | null;

  connect() {
    const button = this.buttonTarget;

    const scheduledAt = new Date();
    const cooldown = Number(button.getAttribute("data-cooldown")) * 1000;
    const label = button.getAttribute("data-label");
    const labelUnit = button.getAttribute("data-label-unit")!;

    const tick = () => {
      const now = new Date();
      const timeElapsed = now.getTime() - scheduledAt.getTime();

      let displaySeconds = 0;
      if (timeElapsed <= cooldown) {
        displaySeconds = Math.round((cooldown - timeElapsed) / 1000);
      }

      if (displaySeconds === 0) {
        button.disabled = false;
        button.textContent = label;
        this.animHandle = null;
      } else {
        button.disabled = true;
        button.textContent = labelUnit.replace("%d", String(displaySeconds));
        this.animHandle = requestAnimationFrame(tick);
      }
    };

    this.animHandle = requestAnimationFrame(tick);
  }

  disconnect() {
    if (this.animHandle != null) {
      cancelAnimationFrame(this.animHandle);
      this.animHandle = null;
    }
  }
}
