/**
 * Dispatch a custom event to render bot protection widget
 */
export function dispatchBotProtectionWidgetEventRender() {
  document.dispatchEvent(new CustomEvent("bot-protection-widget:render"));
}

/**
 * Dispatch a custom event to set captcha failed
 */
export function dispatchBotProtectionWidgetEventUndoRender() {
  document.dispatchEvent(new CustomEvent("bot-protection-widget:undo-render"));
}
