import { DateTime } from "luxon";

// Ref: https://tc39.es/ecma402/#sec-datetimeformat-abstracts
const dateTimeWithTimezoneFormatOption = {
  year: "numeric" as const,
  month: "short" as const,
  day: "numeric" as const,
  hour: "numeric" as const,
  minute: "numeric" as const,
  second: "numeric" as const,
  timeZoneName: "longOffset" as const,
};

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
  return datetime.toLocaleString(dateTimeWithTimezoneFormatOption);
}
