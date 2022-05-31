import {
  intlDateTimeFormatIsSupported,
  intlRelativeTimeFormatIsSupported,
} from "./feature";
import { DateTime } from "luxon";
import { Controller } from "@hotwired/stimulus";

// In order to be backward compatible,
// <span data-date> is equivalent to
// <span data-date data-date-type="relative" data-date-date-style="medium" data-date-time-style="short">
// To suppress time, specify data-date-time-style as an empty attribute.

const DATE_TYPES = ["relative", "absolute"] as const;
type DateType = typeof DATE_TYPES[number];

const DATE_STYLES = ["full", "long", "medium", "short"] as const;
type DateStyle = typeof DATE_STYLES[number];

const TIME_STYLES = ["full", "long", "medium", "short"] as const;
type TimeStyle = typeof TIME_STYLES[number];

function parseDateType(s: string | null): DateType {
  for (const variant of DATE_TYPES) {
    if (variant === s) {
      return variant;
    }
  }
  return "relative";
}

function parseDateStyle(s: string | null): DateStyle | undefined {
  if (s === "") {
    return undefined;
  }

  for (const variant of DATE_STYLES) {
    if (variant === s) {
      return variant;
    }
  }
  return "medium";
}

function parseTimeStyle(s: string | null): TimeStyle | undefined {
  if (s === "") {
    return undefined;
  }

  for (const variant of TIME_STYLES) {
    if (variant === s) {
      return variant;
    }
  }
  return "short";
}

export class FormatDateRelativeController extends Controller {
  connect() {
    const dateSpans = document.documentElement.querySelectorAll("[data-date]");
    const lang = document.documentElement.lang;

    if (lang == null || lang === "") {
      return;
    }

    const hasAbs = intlDateTimeFormatIsSupported();
    const hasRel = intlRelativeTimeFormatIsSupported();

    for (let i = 0; i < dateSpans.length; i++) {
      const dateSpan = dateSpans[i];
      const rfc3339 = dateSpan.getAttribute("data-date");
      const dateType = parseDateType(dateSpan.getAttribute("data-date-type"));
      const dateStyle = parseDateStyle(
        dateSpan.getAttribute("data-date-date-style")
      );
      const timeStyle = parseTimeStyle(
        dateSpan.getAttribute("data-date-time-style")
      );

      if (typeof rfc3339 === "string") {
        const luxonDatetime = DateTime.fromISO(rfc3339);
        const abs = hasAbs
          ? luxonDatetime.toLocaleString(
              {
                dateStyle,
                timeStyle,
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

        if (dateSpan instanceof HTMLElement) {
          // Display the absolute date time as title (tooltip).
          // This is how GitHub shows date time.
          if (abs != null) {
            dateSpan.title = abs;
          }
        }

        if (dateType === "relative") {
          // Prefer showing relative date time,
          // and fallback to absolute date time.
          if (rel != null) {
            dateSpan.textContent = rel;
          } else if (abs != null) {
            dateSpan.textContent = abs;
          }
        } else {
          if (abs != null) {
            dateSpan.textContent = abs;
          }
        }
      }
    }
  }
}

// There is no way to change the display format of <input type="date">.
// So the comprimise is to make the display format of such value matches that of <input type="date">.
// The display format of <input type="date"> is in browser locale for Safari and Firefox.
// For Chrome, the display format is somehow arbitrary :(
export class FormatInputDateController extends Controller {
  static targets = ["inputDate"];

  declare inputDateTarget: HTMLSpanElement;

  connect() {
    const hasAbs = intlDateTimeFormatIsSupported();
    if (!hasAbs) {
      return;
    }

    const dateSpan = this.inputDateTarget;
    const rfc3339 = dateSpan.getAttribute("data-input-date-value");
    if (typeof rfc3339 === "string") {
      const jsDate = new Date(rfc3339);
      if (!isNaN(jsDate.getTime())) {
        dateSpan.textContent = new Intl.DateTimeFormat().format(jsDate);
      }
    }
  }
}
