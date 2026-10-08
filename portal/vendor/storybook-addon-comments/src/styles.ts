const COMMENTS_FIELD_CSS = `
#storybook-panel-root form input:not([type="checkbox"]),
#storybook-panel-root form textarea {
  background: #fff !important;
  background-color: #fff !important;
  color: #000 !important;
  -webkit-text-fill-color: #000 !important;
}
#storybook-panel-root form input:not([type="checkbox"])::placeholder,
#storybook-panel-root form textarea::placeholder {
  color: #666 !important;
  -webkit-text-fill-color: #666 !important;
  opacity: 1;
}
[data-oursky-comments] input[type="checkbox"] {
  appearance: none;
  -webkit-appearance: none;
  width: 14px;
  height: 14px;
  margin: 0;
  padding: 0;
  background: #fff !important;
  background-color: #fff !important;
  border: 1px solid #cbd5e1;
  border-radius: 3px;
  color-scheme: light;
  cursor: pointer;
  flex-shrink: 0;
}
[data-oursky-comments] input[type="checkbox"]:checked {
  background: #0d9488 !important;
  background-color: #0d9488 !important;
  border-color: #0d9488;
}
`;

export function ensureCommentsStyles() {
  if (typeof document === "undefined") return;
  if (document.getElementById("oursky-comments-field-css")) return;
  const style = document.createElement("style");
  style.id = "oursky-comments-field-css";
  style.textContent = COMMENTS_FIELD_CSS;
  document.head.appendChild(style);
}
