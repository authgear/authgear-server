// These helpers convert between plain text typed by a user and an ICU
// MessageFormat pattern, because translation.json values are parsed as
// message patterns by the server (github.com/iawaknahc/gomessageformat).
//
// That library implements ApostropheMode DOUBLE_REQUIRED, which is stricter
// than ICU's default DOUBLE_OPTIONAL:
//
//   - A single ' ALWAYS opens quoted literal text, whatever character
//     follows it, so a literal apostrophe must always be written as ''.
//   - Quoted literal text runs until the next single ', and '' inside it is
//     a literal apostrophe that does not close the quoting.
//   - { and } are syntax anywhere in a pattern: an unmatched { is a parse
//     error, and {word} is silently substituted as an argument.
//   - An unterminated quote is a parse error, so a value that ends up with
//     a stray ' breaks the whole message.
//
// The escaping is done in a single left-to-right pass that tracks whether a
// quoted run is open. Escaping apostrophes and quoting the syntax characters
// as two independent passes does NOT work: the quote closing one run fuses
// with an adjacent doubled apostrophe, and the parser resolves '' first.
// For example "{}" would become "'{''}'", which renders as "{'}".

// Characters that must appear inside a quoted run to be taken literally.
//
// # is deliberately excluded. It is only syntax inside a plural or
// selectordinal clause, and these values are whole messages rather than
// clause bodies, so a bare # already renders as itself. Add it here if these
// helpers are ever reused for text embedded in a plural clause.
const SYNTAX = "{}";

/**
 * Escape plain text so that it renders as itself when parsed as a message
 * pattern. Apply this before writing a plain-text value to translation.json.
 */
export function escapeMessageFormatText(text: string): string {
  let out = "";
  let quoted = false;
  for (const ch of text) {
    if (ch === "'") {
      // A doubled apostrophe is a literal apostrophe whether or not a quoted
      // run is open, and it does not change the quoting state.
      out += "''";
      continue;
    }
    if (SYNTAX.includes(ch) && !quoted) {
      out += "'";
      quoted = true;
    }
    out += ch;
  }
  // Once a quoted run is open it is left open, because every remaining
  // character is literal inside it anyway. Close it at the end.
  if (quoted) {
    out += "'";
  }
  return out;
}

/**
 * Recover the plain text that a message pattern renders as. Apply this when
 * reading a value out of translation.json for display in an input.
 *
 * This resolves quoting only, so it also reads values written by earlier
 * versions of escapeMessageFormatText, which escaped apostrophes but left
 * { and } bare.
 */
export function unescapeMessageFormatText(text: string): string {
  let out = "";
  let i = 0;
  while (i < text.length) {
    if (text[i] !== "'") {
      out += text[i];
      i += 1;
      continue;
    }
    if (text[i + 1] === "'") {
      out += "'";
      i += 2;
      continue;
    }
    // A lone ' opens a quoted run and contributes nothing itself.
    i += 1;
    while (i < text.length) {
      if (text[i] === "'") {
        if (text[i + 1] === "'") {
          out += "'";
          i += 2;
          continue;
        }
        i += 1; // closing quote
        break;
      }
      out += text[i];
      i += 1;
    }
    // An unterminated quoted run extends to the end of the string. The server
    // rejects such a pattern, but being lenient here shows the user something
    // sensible rather than nothing.
  }
  return out;
}
