// Legacy fallback: only reliable outside modal dialogs — Radix Dialog's focus
// trap immediately re-steals focus from the body-level textarea, losing the
// selection before execCommand runs. Kept only for the rare browser context
// without the async Clipboard API (e.g. plain-http non-localhost origins).
function copyToClipboardLegacy(str: string): void {
  const el = document.createElement("textarea");
  el.value = str;
  // Set non-editable to avoid focus and move outside of view
  el.setAttribute("readonly", "");
  el.setAttribute("style", "position: absolute; left: -9999px");
  document.body.appendChild(el);
  // Select text inside element
  el.select();
  el.setSelectionRange(0, el.value.length); // for mobile device

  document.execCommand("copy");
  // Remove temporary element
  document.body.removeChild(el);
}

export function copyToClipboard(str: string): void {
  // lib.dom types navigator.clipboard as always present, but it is actually
  // undefined in insecure contexts — hence the cast before the null check.
  const clipboard = navigator.clipboard as Clipboard | undefined;
  if (clipboard != null) {
    clipboard.writeText(str).catch(() => {
      copyToClipboardLegacy(str);
    });
    return;
  }
  copyToClipboardLegacy(str);
}
