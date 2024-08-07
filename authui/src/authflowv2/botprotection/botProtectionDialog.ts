import { Controller } from "@hotwired/stimulus";
import {
  dispatchBotProtectionWidgetEventRender,
  dispatchBotProtectionWidgetEventUndoRender,
} from "./botProtectionWidget";
import { dispatchDialogClose, dispatchDialogOpen } from "../dialog";

const DIALOG_ID = "bot-protection-dialog";

/**
 * Dispatch a custom event to set captcha dialog open
 */
export function dispatchBotProtectionDialogOpen() {
  dispatchDialogOpen(DIALOG_ID);
}

/**
 * Dispatch a custom event to set captcha dialog close
 */
export function dispatchBotProtectionDialogClose() {
  dispatchDialogClose(DIALOG_ID);
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
    document.addEventListener(`dialog-${DIALOG_ID}:open`, this.open);
    document.addEventListener(`dialog-${DIALOG_ID}:close`, this.close);
  }

  disconnect() {
    document.removeEventListener(`dialog-${DIALOG_ID}:open`, this.open);
    document.removeEventListener(`dialog-${DIALOG_ID}:close`, this.close);
  }
}
