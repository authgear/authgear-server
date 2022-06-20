import { Controller } from "@hotwired/stimulus";

export class ResendButtonController extends Controller {
  static values = {
    cooldown: Number,
    label: String,
    labelUnit: String,
  };

  declare cooldownValue: number;
  declare labelValue: string;
  declare labelUnitValue: string;
  declare animHandle: number | null;

  connect() {
    const button = this.element as HTMLButtonElement;

    const scheduledAt = new Date();
    const cooldown = this.cooldownValue * 1000;
    const label = this.labelValue;
    const labelUnit = this.labelUnitValue;

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
