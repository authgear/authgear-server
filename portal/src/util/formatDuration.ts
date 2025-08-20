import { Duration } from "luxon";

export function formatSeconds(
  locale: string,
  numberOfSeconds: number
): string | null {
  if (numberOfSeconds < 0) {
    return null;
  }

  const duration = Duration.fromObject({
    seconds: numberOfSeconds,
  }).reconfigure({
    locale,
  });

  // Special case: "0 seconds".
  if (numberOfSeconds === 0) {
    return duration.shiftTo("second").toHuman({
      showZeros: true,
      listStyle: "narrow",
    });
  }

  return duration.shiftTo("day", "hour", "minute", "second").toHuman({
    showZeros: false,
    listStyle: "narrow",
  });
}
