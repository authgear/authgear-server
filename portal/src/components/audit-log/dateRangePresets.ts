import { DateTime } from "luxon";
import { formatDateOnly } from "../../util/formatDateOnly";

export type AuditLogDateRangePresetKey =
  | "today"
  | "last7Days"
  | "last30Days"
  | "custom";

export const AUDIT_LOG_DATE_RANGE_PRESET_ORDER: AuditLogDateRangePresetKey[] = [
  "today",
  "last7Days",
  "last30Days",
  "custom",
];

function startOfDay(date: Date): Date {
  return DateTime.fromJSDate(date).startOf("day").toJSDate();
}

function isSameDay(a: Date | null, b: Date | null): boolean {
  if (a == null || b == null) {
    return a === b;
  }
  return DateTime.fromJSDate(a).hasSame(DateTime.fromJSDate(b), "day");
}

function clampFromDate(from: Date, minDate?: Date): Date {
  if (minDate == null) {
    return from;
  }
  const min = startOfDay(minDate);
  return from < min ? min : from;
}

export function getPresetDateRange(
  preset: Exclude<AuditLogDateRangePresetKey, "custom">,
  referenceDate: Date,
  minDate?: Date
): { from: Date; to: Date } {
  const to = startOfDay(referenceDate);
  const toDateTime = DateTime.fromJSDate(to);
  let from: Date;

  switch (preset) {
    case "today":
      from = to;
      break;
    case "last7Days":
      from = toDateTime.minus({ days: 6 }).toJSDate();
      break;
    case "last30Days":
      from = toDateTime.minus({ days: 29 }).toJSDate();
      break;
  }

  return {
    from: clampFromDate(from, minDate),
    to,
  };
}

export function detectDateRangePreset(
  rangeFrom: Date | null,
  rangeTo: Date | null,
  referenceDate: Date,
  minDate?: Date
): AuditLogDateRangePresetKey {
  if (rangeFrom == null && rangeTo == null) {
    return "today";
  }

  for (const preset of ["today", "last7Days", "last30Days"] as const) {
    const expected = getPresetDateRange(preset, referenceDate, minDate);
    if (
      isSameDay(rangeFrom, expected.from) &&
      isSameDay(rangeTo, expected.to)
    ) {
      return preset;
    }
  }

  return "custom";
}

export function getInitialAuditLogDateRange(
  queryFrom: string | null,
  queryTo: string | null,
  queryLastUpdatedAt: string | null
): {
  preset: AuditLogDateRangePresetKey;
  rangeFrom: Date | null;
  rangeTo: Date | null;
} {
  const referenceDate =
    queryLastUpdatedAt != null
      ? new Date(Number(queryLastUpdatedAt))
      : new Date();
  const fromParam =
    queryFrom != null && queryFrom !== "" ? new Date(queryFrom) : null;
  const toParam =
    queryTo != null && queryTo !== "" ? new Date(queryTo) : null;
  const preset = detectDateRangePreset(fromParam, toParam, referenceDate);

  if (preset === "custom") {
    return { preset, rangeFrom: fromParam, rangeTo: toParam };
  }

  const range = getPresetDateRange(preset, referenceDate);
  return { preset, rangeFrom: range.from, rangeTo: range.to };
}

export function formatCustomDateRangeLabel(
  locale: string,
  rangeFrom: Date | null,
  rangeTo: Date | null
): string | null {
  if (rangeFrom == null || rangeTo == null) {
    return null;
  }

  const fromDateTime = DateTime.fromJSDate(rangeFrom).setLocale(locale);
  const toDateTime = DateTime.fromJSDate(rangeTo);
  const fromLabel = formatDateOnly(locale, rangeFrom);
  const toLabel = formatDateOnly(locale, rangeTo);
  if (fromLabel == null || toLabel == null) {
    return null;
  }

  if (isSameDay(rangeFrom, rangeTo)) {
    return fromLabel;
  }

  if (fromDateTime.hasSame(toDateTime, "year")) {
    const fromMonthDay = fromDateTime.toLocaleString({
      month: "short",
      day: "numeric",
    });
    return `${fromMonthDay} – ${toLabel}`;
  }

  return `${fromLabel} – ${toLabel}`;
}
