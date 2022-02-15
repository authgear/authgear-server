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
