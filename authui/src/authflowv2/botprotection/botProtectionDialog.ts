import { Controller } from "@hotwired/stimulus";
import {
  dispatchBotProtectionWidgetEventRender,
  dispatchBotProtectionWidgetEventUndoRender,
} from "./botProtectionWidget";
import { dispatchDialogClose, dispatchDialogOpen } from "../dialog";

// Assume globally only have ONE single dialog
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
  onOpenEnd = (e: Event) => {
    if (!(e instanceof CustomEvent)) {
      return;
    }
    if (e.detail.id !== DIALOG_ID) {
      // event targets other dialog
      return;
    }
    dispatchBotProtectionWidgetEventRender();
  };

  onCloseEnd = (e: Event) => {
    if (!(e instanceof CustomEvent)) {
      return;
    }
    if (e.detail.id !== DIALOG_ID) {
      // event targets other dialog
      return;
    }
    dispatchBotProtectionWidgetEventUndoRender();
  };

  connect() {
    if (this.element.id !== DIALOG_ID) {
      console.error(`bot-protection-dialog must have id="${DIALOG_ID}"`);
      return;
    }
    document.addEventListener(`dialog:opened`, this.onOpenEnd);
    document.addEventListener(`dialog:closed`, this.onCloseEnd);
  }

  disconnect() {
    document.removeEventListener(`dialog:opened`, this.onOpenEnd);
    document.removeEventListener(`dialog:closed`, this.onCloseEnd);
  }
}
