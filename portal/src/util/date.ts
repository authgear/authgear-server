import { DateTime } from "luxon";

export function isoWeekLabel(isoDate: string): string {
  let label = isoDate;
  const luxonDate = DateTime.fromISO(isoDate, {
    zone: "UTC",
  });

  const iosWeek = luxonDate.toISOWeekDate(); //=> '1982-W21-2'
  const parts = iosWeek.split("-");
  if (parts.length === 3) {
    label = `${label} (${parts[1]})`;
  }

  return label;
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
