import { Controller } from "@hotwired/stimulus";
import { dispatchBotProtectionWidgetEventRender } from "./botProtectionWidget";

/**
 * Dispatch a custom event to set captcha dialog open
 */
export function dispatchBotProtectionDialogOpen() {
  document.dispatchEvent(new CustomEvent("bot-protection-dialog:open"));
}

/**
 * Dispatch a custom event to set captcha dialog close
 */
export function dispatchBotProtectionDialogClose() {
  document.dispatchEvent(new CustomEvent("bot-protection-dialog:close"));
}

/**
 * Controller for bot protection dialog display
 *
 * Expected usage:
 * - Add `data-controller="bot-protection-dialog"` to a dialog
 */
export class BotProtectionDialogController extends Controller {
  open = (e: Event) => {
    if (!(e instanceof CustomEvent)) {
      throw new Error("Unexpected non-CustomEvent");
    }
    dispatchBotProtectionWidgetEventRender();
    this.element.classList.add("open");
  };

  close = () => {
    this.element.classList.remove("open");
  };

  connect() {
    document.addEventListener("bot-protection-dialog:open", this.open);
    document.addEventListener("bot-protection-dialog:close", this.close);
  }

  disconnect() {
    document.removeEventListener("bot-protection-dialog:open", this.open);
    document.removeEventListener("bot-protection-dialog:close", this.close);
  }
}
