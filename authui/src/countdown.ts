import { Controller } from "@hotwired/stimulus";
import { Duration } from "luxon";

export class CountdownController extends Controller {
  static targets = ["button"];
  static values = {
    cooldown: Number,
    cooldownUntil: String,
    label: String,
    labelUnit: String,
    format: String,
  };

  declare readonly buttonTarget: HTMLButtonElement;

  declare cooldownValue?: number;
  declare cooldownUntilValue?: string;
  declare labelValue: string;
  declare labelUnitValue: string;
  declare formatValue?: string;

  animHandle: number | null = null;

  connect() {
    this.validateValues();

    const button = this.buttonTarget;

    const scheduledAt = new Date();
    const cooldown = this.cooldownValue
      ? this.cooldownValue * 1000
      : new Date(this.cooldownUntilValue!).getTime() - scheduledAt.getTime();
    const label = this.labelValue;
    const labelUnit = this.labelUnitValue;
    const format = this.formatValue || "mm:ss";

    const tick = () => {
      const now = new Date();
      const timeElapsed = now.getTime() - scheduledAt.getTime();

      let seconds = 0;
      if (timeElapsed <= cooldown) {
        seconds = Math.round((cooldown - timeElapsed) / 1000);
      }
      const duration = Duration.fromObject({ seconds });
      const formatted = duration.toFormat(format);

      if (seconds === 0) {
        button.disabled = false;
        button.textContent = label;
        this.animHandle = null;
      } else {
        button.disabled = true;
        button.textContent = labelUnit.replace("%s", formatted);
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

  private validateValues() {
    if (this.cooldownValue && isNaN(this.cooldownValue)) {
      throw new Error("cooldown must be a valid number");
    }

    if (
      this.cooldownUntilValue &&
      isNaN(new Date(this.cooldownUntilValue).getTime())
    ) {
      throw new Error("cooldownUntil must be a valid date");
    }
  }
}
