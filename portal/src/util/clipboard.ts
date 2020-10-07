export function copyToClipboard(str: string): void {
  const el = document.createElement("textarea");
  el.value = str;
  // Set non-editable to avoid focus and move outside of view
  el.setAttribute("readonly", "");
  el.setAttribute("style", "position: absolute; left: -9999px");
  document.body.appendChild(el);
  // Select text inside element
  el.select();
  el.setSelectionRange(0, 100); // for mobile device
  document.execCommand("copy");
  // Remove temporary element
  document.body.removeChild(el);
}
