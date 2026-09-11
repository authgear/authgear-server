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

// rangeFrom and rangeTo are both inclusive instants at the minute granularity
// of the date-time picker. A whole local day therefore runs from 00:00 to
// 23:59; toExclusiveRangeTo turns the latter into the exclusive bound the
// backend expects.

function startOfDay(date: Date): Date {
  return DateTime.fromJSDate(date).startOf("day").toJSDate();
}

function endOfDay(date: Date): Date {
  return DateTime.fromJSDate(date).endOf("day").startOf("minute").toJSDate();
}

function isStartOfDay(date: Date): boolean {
  return date.getTime() === startOfDay(date).getTime();
}

function isEndOfDay(date: Date): boolean {
  return date.getTime() === endOfDay(date).getTime();
}

const DATE_ONLY_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

// The from/to query params are ISO datetimes with offset (see
// serializeDateRangeSearchParam). Before the picker had a time component they
// were local calendar dates written with DateTime.toISODate(); those still
// live in bookmarks and must be read back in the local zone as whole inclusive
// days. `new Date("YYYY-MM-DD")` would parse as UTC midnight, which is the
// previous calendar day in every UTC-negative zone.
export function parseDateRangeSearchParam(
  value: string | null,
  bound: "from" | "to"
): Date | null {
  if (value == null || value === "") {
    return null;
  }
  const date = DateTime.fromISO(value).toJSDate();
  if (isNaN(date.getTime())) {
    return null;
  }
  if (DATE_ONLY_PATTERN.test(value)) {
    return bound === "to" ? endOfDay(date) : startOfDay(date);
  }
  return date;
}

export function serializeDateRangeSearchParam(date: Date | null): string {
  if (date == null) {
    return "";
  }
  return DateTime.fromJSDate(date).toISO({ suppressMilliseconds: true });
}

// The backend filters with `created_at < rangeTo`. rangeTo is inclusive at
// minute granularity, so the exclusive bound is the start of the next minute.
export function toExclusiveRangeTo(rangeTo: Date): Date {
  return DateTime.fromJSDate(rangeTo)
    .plus({ minutes: 1 })
    .startOf("minute")
    .toJSDate();
}

function isSameInstant(a: Date | null, b: Date | null): boolean {
  if (a == null || b == null) {
    return a === b;
  }
  return a.getTime() === b.getTime();
}

function isSameDay(a: Date, b: Date): boolean {
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
  const to = endOfDay(referenceDate);
  const startOfReferenceDay = DateTime.fromJSDate(startOfDay(referenceDate));
  let from: Date;

  switch (preset) {
    case "today":
      from = startOfReferenceDay.toJSDate();
      break;
    case "last7Days":
      from = startOfReferenceDay.minus({ days: 6 }).toJSDate();
      break;
    case "last30Days":
      from = startOfReferenceDay.minus({ days: 29 }).toJSDate();
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
      isSameInstant(rangeFrom, expected.from) &&
      isSameInstant(rangeTo, expected.to)
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
  const fromParam = parseDateRangeSearchParam(queryFrom, "from");
  const toParam = parseDateRangeSearchParam(queryTo, "to");
  const preset = detectDateRangePreset(fromParam, toParam, referenceDate);

  if (preset === "custom") {
    return { preset, rangeFrom: fromParam, rangeTo: toParam };
  }

  const range = getPresetDateRange(preset, referenceDate);
  return { preset, rangeFrom: range.from, rangeTo: range.to };
}

const dateTimeLabelFormat = {
  month: "short" as const,
  day: "numeric" as const,
  hour: "numeric" as const,
  minute: "numeric" as const,
};

const dateTimeWithYearLabelFormat = {
  ...dateTimeLabelFormat,
  year: "numeric" as const,
};

function formatWholeDaysLabel(
  locale: string,
  rangeFrom: Date,
  rangeTo: Date
): string | null {
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

function formatDateTimesLabel(
  locale: string,
  rangeFrom: Date,
  rangeTo: Date
): string {
  const fromDateTime = DateTime.fromJSDate(rangeFrom).setLocale(locale);
  const toDateTime = DateTime.fromJSDate(rangeTo).setLocale(locale);
  const fromLabel = fromDateTime.toLocaleString(
    fromDateTime.hasSame(toDateTime, "year")
      ? dateTimeLabelFormat
      : dateTimeWithYearLabelFormat
  );
  const toLabel = toDateTime.toLocaleString(dateTimeWithYearLabelFormat);
  return `${fromLabel} – ${toLabel}`;
}

export function formatCustomDateRangeLabel(
  locale: string,
  rangeFrom: Date | null,
  rangeTo: Date | null
): string | null {
  if (rangeFrom == null || rangeTo == null) {
    return null;
  }

  // Only show the times when the range does not cover whole days, so preset
  // style ranges stay as compact as before.
  if (isStartOfDay(rangeFrom) && isEndOfDay(rangeTo)) {
    return formatWholeDaysLabel(locale, rangeFrom, rangeTo);
  }
  return formatDateTimesLabel(locale, rangeFrom, rangeTo);
}
