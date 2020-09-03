import { DateTime, DateTimeFormatOptions } from "luxon";

export function formatDatetime(
  locale: string,
  date: Date | string | null,
  format: DateTimeFormatOptions = DateTime.DATETIME_SHORT
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
  return datetime.toLocaleString(format);
}
