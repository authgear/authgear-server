import { DateTime } from "luxon";

export function parseBirthdate(str: string): Date | undefined {
  try {
    const datetime = DateTime.fromFormat(str, "yyyy-MM-dd", {
      zone: "Etc/UTC",
    });
    const s = datetime.toFormat("yyyy-MM-dd");
    if (s !== str) {
      return undefined;
    }
    return datetime.toJSDate();
  } catch {}
  return undefined;
}

export function toBirthdate(date: Date): string | undefined {
  if (isNaN(date.getTime())) {
    return undefined;
  }
  return DateTime.fromJSDate(date).toFormat("yyyy-MM-dd");
}
