import {
  intlDateTimeFormatIsSupported,
  intlRelativeTimeFormatIsSupported,
} from "./feature";
import { DateTime } from "luxon";

export function formatDateRelative() {
  const dateSpans = document.querySelectorAll("[data-date]");
  const lang = document.documentElement.lang;

  if (lang == null || lang === "") {
    return;
  }

  const hasAbs = intlDateTimeFormatIsSupported();
  const hasRel = intlRelativeTimeFormatIsSupported();

  for (let i = 0; i < dateSpans.length; i++) {
    const dateSpan = dateSpans[i];
    const rfc3339 = dateSpan.getAttribute("data-date");
    if (typeof rfc3339 === "string") {
      const luxonDatetime = DateTime.fromISO(rfc3339);
      const abs = hasAbs
        ? luxonDatetime.toLocaleString(
            {
              dateStyle: "medium",
              timeStyle: "short",
            },
            {
              locale: lang,
            }
          )
        : null;
      const rel = hasRel
        ? luxonDatetime.toRelative({
            locale: lang,
          })
        : null;

      // Store the original textContent.
      const textContent = dateSpan.textContent;
      if (textContent != null) {
        dateSpan.setAttribute("data-original-text-content", textContent);
      }

      if (dateSpan instanceof HTMLElement) {
        // Display the absolute date time as title (tooltip).
        // This is how GitHub shows date time.
        if (abs != null) {
          dateSpan.title = abs;
        }
      }

      // Prefer showing relative date time,
      // and fallback to absolute date time.
      if (rel != null) {
        dateSpan.textContent = rel;
      } else if (abs != null) {
        dateSpan.textContent = abs;
      }
    }
  }
}

// There is no way to change the display format of <input type="date">.
// So the comprimise is to make the display format of such value matches that of <input type="date">.
// The display format of <input type="date"> is in browser locale for Safari and Firefox.
// For Chrome, the display format is somehow arbitrary :(
export function formatInputDate() {
  const hasAbs = intlDateTimeFormatIsSupported();
  if (!hasAbs) {
    return;
  }

  const dateSpans = document.querySelectorAll("[data-input-date-value]");
  for (let i = 0; i < dateSpans.length; i++) {
    const dateSpan = dateSpans[i];
    const rfc3339 = dateSpan.getAttribute("data-input-date-value");
    if (typeof rfc3339 === "string") {
      const jsDate = new Date(rfc3339);
      if (!isNaN(jsDate.getTime())) {
        // Store the original textContent.
        const textContent = dateSpan.textContent;
        if (textContent != null) {
          dateSpan.setAttribute("data-original-text-content", textContent);
        }

        dateSpan.textContent = new Intl.DateTimeFormat().format(jsDate);
      }
    }
  }
}
