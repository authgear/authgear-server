import { DateTime } from "luxon";

const dateTimeFormatOption = {
  year: "numeric" as const,
  month: "short" as const,
  day: "numeric" as const,
  hour: "numeric" as const,
  minute: "numeric" as const,
  second: "numeric" as const,
};

const dateTimeWithoutSecondsFormatOption = {
  month: "short" as const,
  day: "numeric" as const,
  hour: "numeric" as const,
  minute: "numeric" as const,
};

const dateTimeWithYearWithoutSecondsFormatOption = {
  ...dateTimeWithoutSecondsFormatOption,
  year: "numeric" as const,
};

// Ref: https://tc39.es/ecma402/#sec-datetimeformat-abstracts
const dateTimeWithTimezoneFormatOption = {
  ...dateTimeFormatOption,
  timeZoneName: "longOffset" as const,
};

function toDateTime(date: Date | string | null): DateTime | null {
  if (date instanceof Date) {
    return DateTime.fromJSDate(date);
  }
  if (typeof date === "string") {
    return DateTime.fromISO(date);
  }
  return null;
}

export function formatDatetimeWithoutTimezone(
  locale: string,
  date: Date | string | null
): string | null {
  const datetime = toDateTime(date);
  if (datetime == null) {
    return null;
  }

  return datetime.setLocale(locale).toLocaleString(dateTimeFormatOption);
}

export function formatCustomDateRangeLabel(
  rangeFrom: Date | null,
  rangeTo: Date | null
): string | undefined {
  // Match DateTimePicker / Authgear datetime display (en-US 12-hour AM/PM).
  const locale = "en-US";
  const fromLabel =
    rangeFrom != null
      ? toDateTime(rangeFrom)
          ?.setLocale(locale)
          .toLocaleString(dateTimeWithoutSecondsFormatOption) ?? null
      : null;
  const toLabel =
    rangeTo != null
      ? toDateTime(rangeTo)
          ?.setLocale(locale)
          .toLocaleString(dateTimeWithYearWithoutSecondsFormatOption) ?? null
      : null;

  if (fromLabel != null && toLabel != null) {
    return `${fromLabel} - ${toLabel}`;
  }
  if (fromLabel != null) {
    return fromLabel;
  }
  if (toLabel != null) {
    return toLabel;
  }
  return undefined;
}

export function formatDatetime(
  locale: string,
  date: Date | string | null
): string | null {
  let datetime;

  if (date instanceof Date) {
    datetime = DateTime.fromJSDate(date);
  } else if (typeof date === "string") {
    datetime = DateTime.fromISO(date);
  }

  if (datetime == null) {
    return null;
  }

  datetime = datetime.setLocale(locale);
  return datetime
    .toLocaleString(dateTimeWithTimezoneFormatOption)
    .replace("GMT+", "UTC+");
}
