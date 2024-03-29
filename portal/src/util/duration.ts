// ref: https://cs.opensource.google/go/go/+/refs/tags/go1.20.3:src/time/format.go;l=1589

const units = {
  ns: 1e-9,
  us: 1e-6,
  µs: 1e-6,
  μs: 1e-6,
  ms: 1e-3,
  s: 1,
  m: 60,
  h: 60 * 60,
} as const;

type DurationUnit = keyof typeof units;

const partRegex = new RegExp(
  `([0-9]*(?:\\.[0-9]*)?)(${Object.keys(units).join("|")})`,
  "g"
);

const durationRegex = new RegExp(`^[-|+]?((${partRegex.source})+|0)$`, "g");

export function parseDuration(s: string): number {
  if (!s.match(durationRegex)) {
    throw new Error("Invalid duration string");
  }
  let sign = 1;
  if (s[0] === "-" || s[1] === "+") {
    sign = s[0] === "-" ? -1 : 1;
    s = s.slice(1);
  }
  if (s === "0") {
    return 0;
  }

  let seconds = 0;
  for (const match of s.matchAll(partRegex)) {
    const [, num, unit] = match;
    const value = Number(num) * units[unit as DurationUnit];
    seconds += value;
  }
  return sign * seconds;
}

export function formatDuration(quantity: number, unit: DurationUnit): string {
  return quantity.toString() + unit;
}

export function formatOptionalDuration(
  quantity: number | undefined,
  unit: DurationUnit
): string | undefined {
  if (quantity == null) {
    return undefined;
  }
  return formatDuration(quantity, unit);
}
