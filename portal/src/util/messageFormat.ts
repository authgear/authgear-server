// ICU MessageFormat treats a single quote (') as the start of a quoted
// literal section. A plain-text value with an odd number of quotes (e.g.
// "O'Brien") therefore fails to parse as a message pattern. Doubling every
// quote escapes it to a literal quote, which parses correctly.
export function escapeMessageFormatText(text: string): string {
  return text.replace(/'/g, "''");
}

export function unescapeMessageFormatText(text: string): string {
  return text.replace(/''/g, "'");
}
