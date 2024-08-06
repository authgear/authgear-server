import { Controller } from "@hotwired/stimulus";
import {
  dispatchBotProtectionWidgetEventRender,
  dispatchBotProtectionWidgetEventUndoRender,
} from "./botProtectionWidget";

const DIALOG_ID = "bot-protection-dialog";
const BOT_PROTECTION_DIALOG_OPEN_EVENT = `dialog-${DIALOG_ID}:open`;
const BOT_PROTECTION_DIALOG_CLOSE_EVENT = `dialog-${DIALOG_ID}:close`;

/**
 * Dispatch a custom event to set captcha dialog open
 */
export function dispatchBotProtectionDialogOpen() {
  document.dispatchEvent(new CustomEvent(BOT_PROTECTION_DIALOG_OPEN_EVENT));
}

/**
 * Dispatch a custom event to set captcha dialog close
 */
export function dispatchBotProtectionDialogClose() {
  document.dispatchEvent(new CustomEvent(BOT_PROTECTION_DIALOG_CLOSE_EVENT));
}

/**
 * Controller for bot protection dialog display
 *
 * Expected usage:
 * - Add `data-controller="bot-protection-dialog"` to a dialog
 * - Specify id="bot-protection-dialog" to that dialog
 */
export class BotProtectionDialogController extends Controller {
  open = () => {
    dispatchBotProtectionWidgetEventRender();
  };

  close = () => {
    dispatchBotProtectionWidgetEventUndoRender();
  };

  connect() {
    if (this.element.id !== DIALOG_ID) {
      console.error(`bot-protection-dialog must have id="${DIALOG_ID}"`);
      return;
    }
    document.addEventListener(BOT_PROTECTION_DIALOG_OPEN_EVENT, this.open);
    document.addEventListener(BOT_PROTECTION_DIALOG_CLOSE_EVENT, this.close);
  }

  disconnect() {
    document.removeEventListener(BOT_PROTECTION_DIALOG_OPEN_EVENT, this.open);
    document.removeEventListener(BOT_PROTECTION_DIALOG_CLOSE_EVENT, this.close);
  }
}
