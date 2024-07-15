import { DateTime, DateTimeFormatOptions } from "luxon";

export function formatDateOnly(
  locale: string,
  date: Date | string | null,
  format: DateTimeFormatOptions = DateTime.DATE_MED
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
