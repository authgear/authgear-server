export function intlDateTimeFormatIsSupported(): boolean {
  if ("Intl" in window && "DateTimeFormat" in Intl) {
    const testDate = new Date("2006-01-02T03:04:05Z");
    const testFormat = new Intl.DateTimeFormat("en", {
      dateStyle: "long",
      timeStyle: "long",
      timeZone: "UTC",
    } as any);
    const actual = testFormat.format(testDate);
    const expected = "January 2, 2006 at 3:04:05 AM UTC";
    return actual === expected;
  }
  return false;
}

export function intlRelativeTimeFormatIsSupported(): boolean {
  if ("Intl" in window && "RelativeTimeFormat" in Intl) {
    return true;
  }
  return false;
}
