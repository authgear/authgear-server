import {
  intlDateTimeFormatIsSupported,
  intlRelativeTimeFormatIsSupported,
} from "../feature";
import { DateTime, Duration } from "luxon";
import { Controller } from "@hotwired/stimulus";

// In order to be backward compatible,
// <span data-date> is equivalent to
// <span data-date data-date-type="relative" data-date-date-style="medium" data-date-time-style="short">
// To suppress time, specify data-date-time-style as an empty attribute.

const DATE_TYPES = ["relative", "absolute", "luxon-relative-duration"] as const;
type DateType = (typeof DATE_TYPES)[number];

const DATE_STYLES = ["full", "long", "medium", "short"] as const;
type DateStyle = (typeof DATE_STYLES)[number];

const TIME_STYLES = ["full", "long", "medium", "short"] as const;
type TimeStyle = (typeof TIME_STYLES)[number];

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

function formatLuxonRelativeDuration(
  lang: string,
  dt: DateTime,
  base: DateTime
): string {
  let duration = dt.diff(base);
  duration = Duration.fromMillis(
    // Trim to seconds
    Math.trunc(duration.toMillis() / 1000) * 1000,
    {
      locale: lang,
    }
  ).rescale();
  const opts = {
    unitDisplay: "narrow",
    listStyle: "narrow",
    type: "unit",
  } as const;
  return duration.reconfigure({ locale: lang }).toHuman(opts);
}

function getTimezoneFromMetaTag(): string | undefined {
  const metaTag = document.querySelector('meta[name="x-zoneinfo"]');
  const timezone = metaTag?.getAttribute("content");
  return !!timezone ? timezone : undefined;
}

export class FormatDateController extends Controller {
  static values = {
    relativeBase: String,
  };

  declare relativeBaseValue: string;

  private observer: MutationObserver | null = null;

  connect(): void {
    this.render();
    this.observer = new MutationObserver(() => {
      this.observer?.disconnect();
      this.render();
      this.observer?.observe(this.element, { childList: true, subtree: true });
    });
    this.observer.observe(this.element, { childList: true, subtree: true });
  }

  disconnect(): void {
    this.observer?.disconnect();
  }

  relativeBaseValueChanged() {
    this.render();
  }

  // eslint-disable-next-line sonarjs/cognitive-complexity,complexity
  render = () => {
    const lang = document.documentElement.lang;
    const timezone = getTimezoneFromMetaTag();

    if (lang === "") {
      return;
    }

    const hasAbs = intlDateTimeFormatIsSupported();
    const hasRel = intlRelativeTimeFormatIsSupported();
    let relativeBase = DateTime.now();
    if (this.relativeBaseValue) {
      relativeBase = DateTime.fromISO(this.relativeBaseValue);
    }

    const dateSpans = this.element.querySelectorAll("[data-date]");

    for (const dateSpan of dateSpans) {
      const rfc3339 = dateSpan.getAttribute("data-date");
      const dateType = parseDateType(dateSpan.getAttribute("data-date-type"));
      const dateStyle = parseDateStyle(
        dateSpan.getAttribute("data-date-date-style")
      );
      const timeStyle = parseTimeStyle(
        dateSpan.getAttribute("data-date-time-style")
      );

      if (typeof rfc3339 === "string") {
        const luxonDatetime = DateTime.fromISO(rfc3339, { zone: timezone });
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
              base: relativeBase,
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
        } else if (dateType === "luxon-relative-duration") {
          dateSpan.textContent = formatLuxonRelativeDuration(
            lang,
            luxonDatetime,
            relativeBase
          );
        } else {
          if (abs != null) {
            dateSpan.textContent = abs;
          }
        }
      }
    }
  };
}

// There is no way to change the display format of <input type="date">.
// So the comprimise is to make the display format of such value matches that of <input type="date">.
// The display format of <input type="date"> is in browser locale for Safari and Firefox.
// For Chrome, the display format is somehow arbitrary :(
export class FormatInputDateController extends Controller {
  static values = {
    date: String,
  };

  declare dateValue: string;

  connect() {
    const hasAbs = intlDateTimeFormatIsSupported();
    if (!hasAbs) {
      return;
    }

    const dateSpan = this.element as HTMLSpanElement;
    const rfc3339 = this.dateValue;
    if (rfc3339 !== "") {
      const jsDate = new Date(rfc3339);
      if (!isNaN(jsDate.getTime())) {
        dateSpan.textContent = new Intl.DateTimeFormat().format(jsDate);
      }
    }
  }
}
