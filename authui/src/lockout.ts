import { Controller } from "@hotwired/stimulus";
import { DateTime } from "luxon";

export class LockoutController extends Controller {
  static targets = ["locked", "unlocked"];
  static values = {
    lockUntil: String,
    actionButtonSelector: String,
  };

  declare intervalHandle: number | null;
  declare lockUntilValue: string;
  declare actionButtonSelectorValue: string;
  declare lockedTargets: HTMLElement[];
  declare unlockedTargets: HTMLElement[];

  private isLocked: boolean | undefined = undefined;
  private wasActionButtonDisabled: boolean = false;

  stopTick() {
    if (this.intervalHandle != null) {
      window.clearInterval(this.intervalHandle);
      this.intervalHandle = null;
      this.wasActionButtonDisabled = false;
    }
  }

  setupTick() {
    this.isLocked = undefined;
    this.stopTick();
    const el = this.element as HTMLElement;
    const lockUntil = DateTime.fromISO(this.lockUntilValue);
    if (!this.lockUntilValue || !lockUntil.isValid) {
      return;
    }

    const tick = () => {
      const actionButtonEl = el.querySelector(this.actionButtonSelectorValue);
      const now = DateTime.now();

      const newIsLocked = now < lockUntil;
      if (newIsLocked === this.isLocked) {
        return;
      }
      this.isLocked = newIsLocked;
      if (actionButtonEl != null) {
        if (newIsLocked === true) {
          this.wasActionButtonDisabled = actionButtonEl?.getAttribute(
            "disabled"
          )
            ? true
            : false;
          (actionButtonEl as HTMLButtonElement).disabled = true;
        } else {
          (actionButtonEl as HTMLButtonElement).disabled =
            this.wasActionButtonDisabled;
        }
      }

      this.lockedTargets.forEach(
        (el) => ((el as HTMLElement).style.display = newIsLocked ? "" : "none")
      );
      this.unlockedTargets.forEach(
        (el) => ((el as HTMLElement).style.display = newIsLocked ? "none" : "")
      );

      if (!newIsLocked) {
        this.stopTick();
      }
    };
    tick();

    this.intervalHandle = window.setInterval(tick, 100);
  }

  connect(): void {
    this.setupTick();
  }

  disconnect() {
    this.stopTick();
  }
}
