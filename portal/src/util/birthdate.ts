import { DateTime } from "luxon";

export function parseBirthdate(str: string): Date | undefined {
  try {
    const datetime = DateTime.fromFormat(str, "yyyy-MM-dd", {
      zone: "Etc/UTC",
    });
    const s = datetime.toFormat("yyyy-MM-dd");
    // 0001-01-01 is the zero value in golang
    if (s !== str || s === "0001-01-01") {
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
