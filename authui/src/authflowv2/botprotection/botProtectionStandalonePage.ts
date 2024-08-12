import { Controller } from "@hotwired/stimulus";
import {
  dispatchBotProtectionWidgetEventRender,
  dispatchBotProtectionWidgetEventUndoRender,
} from "./botProtectionWidget";

/**
 * Controller for bot protection standalone page `verify_bot_protection.html`
 *  - Manage widget rendering
 *
 * @listens bot-protection-widget:ready-for-render
 * @fires bot-protection-widget:render
 * @fires bot-protection-widget:undo-render
 * Expected usage:
 * - Add `data-controller="bot-protection-standalone-page"` to a `<div>` element
 */
export class BotProtectionStandalonePageController extends Controller {
  onBPWidgetReadyForRender = () => {
    dispatchBotProtectionWidgetEventRender();
  };
  connect() {
    document.addEventListener(
      "bot-protection-widget:ready-for-render",
      this.onBPWidgetReadyForRender
    );
  }

  disconnect() {
    document.removeEventListener(
      "bot-protection-widget:ready-for-render",
      this.onBPWidgetReadyForRender
    );
    dispatchBotProtectionWidgetEventUndoRender();
  }
}
