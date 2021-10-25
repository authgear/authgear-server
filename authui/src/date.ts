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
