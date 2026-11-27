import { DateTime } from "luxon";

export function isoWeekLabels(isoDate: string): string[] {
  const label = isoDate;
  const labels = [label];
  const luxonDate = DateTime.fromISO(isoDate, {
    zone: "UTC",
  });

  const iosWeek = luxonDate.toISOWeekDate(); //=> '1982-W21-2'
  const parts = iosWeek.split("-");
  if (parts.length === 3) {
    labels.push(`(${parts[1]})`);
  }

  return labels;
}

export function monthLabel(isoDate: string): string {
  const luxonDate = DateTime.fromISO(isoDate, {
    zone: "UTC",
  });

  return luxonDate.toLocaleString(
    {
      month: "short",
      year: "numeric",
    },
    {
      // TODO: support other locale
      locale: "en",
    }
  );
}

export function parseDate(isoDate: string): Date {
  return DateTime.fromISO(isoDate, {
    zone: "UTC",
  }).toJSDate();
}

export function addDays(date: Date, days: number): Date {
  const result = new Date(date);
  result.setDate(result.getDate() + days);
  return result;
}

// Date pickers and date labels read a Date in the local time zone, so a UTC
// midnight shows as the previous day west of UTC.
export function utcDateToLocalDate(date: Date): Date {
  return new Date(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate());
}
