import { describe, it, expect } from "@jest/globals";
import { DateTime } from "luxon";
import {
  detectDateRangePreset,
  formatCustomDateRangeLabel,
  getInitialAuditLogDateRange,
  getPresetDateRange,
  parseDateRangeSearchParam,
  serializeDateRangeSearchParam,
  toExclusiveRangeTo,
} from "./dateRangePresets";

function local(
  year: number,
  month: number,
  day: number,
  hour = 0,
  minute = 0,
  second = 0
): Date {
  return DateTime.local(year, month, day, hour, minute, second).toJSDate();
}

describe("getPresetDateRange", () => {
  it("spans local 00:00 to local 23:59 of the reference day for today", () => {
    const { from, to } = getPresetDateRange(
      "today",
      local(2026, 9, 10, 15, 42, 7)
    );
    expect(from).toEqual(local(2026, 9, 10, 0, 0));
    expect(to).toEqual(local(2026, 9, 10, 23, 59));
  });

  it("starts 6 days earlier at 00:00 for last7Days", () => {
    const { from, to } = getPresetDateRange(
      "last7Days",
      local(2026, 9, 10, 15, 42, 7)
    );
    expect(from).toEqual(local(2026, 9, 4, 0, 0));
    expect(to).toEqual(local(2026, 9, 10, 23, 59));
  });
});

describe("toExclusiveRangeTo", () => {
  it("advances an inclusive minute to the start of the next minute", () => {
    expect(toExclusiveRangeTo(local(2026, 9, 1, 14, 30))).toEqual(
      local(2026, 9, 1, 14, 31)
    );
  });

  it("drops seconds so an odd instant still covers its whole minute", () => {
    expect(toExclusiveRangeTo(local(2026, 9, 1, 14, 30, 30))).toEqual(
      local(2026, 9, 1, 14, 31)
    );
  });

  it("turns a preset end of day into the next local midnight", () => {
    const { to } = getPresetDateRange("today", local(2026, 9, 10, 12));
    expect(toExclusiveRangeTo(to)).toEqual(local(2026, 9, 11, 0, 0));
  });
});

describe("parseDateRangeSearchParam", () => {
  it("reads a legacy date-only from as local start of day", () => {
    expect(parseDateRangeSearchParam("2026-09-01", "from")).toEqual(
      local(2026, 9, 1, 0, 0)
    );
  });

  it("reads a legacy date-only to as local 23:59 so the day stays inclusive", () => {
    expect(parseDateRangeSearchParam("2026-09-03", "to")).toEqual(
      local(2026, 9, 3, 23, 59)
    );
  });

  it("reads an ISO datetime with offset as that exact instant", () => {
    expect(
      parseDateRangeSearchParam("2026-09-01T14:30:00+08:00", "to")
    ).toEqual(new Date("2026-09-01T06:30:00Z"));
  });

  it("returns null for empty or unparseable values", () => {
    expect(parseDateRangeSearchParam(null, "from")).toBeNull();
    expect(parseDateRangeSearchParam("", "from")).toBeNull();
    expect(parseDateRangeSearchParam("not-a-date", "to")).toBeNull();
  });
});

describe("serializeDateRangeSearchParam", () => {
  it("writes an ISO datetime with offset and no milliseconds", () => {
    const value = serializeDateRangeSearchParam(local(2026, 9, 1, 14, 30));
    expect(value).toMatch(/^2026-09-01T14:30:00[+-]\d{2}:\d{2}$/);
  });

  it("writes an empty string for null", () => {
    expect(serializeDateRangeSearchParam(null)).toEqual("");
  });

  it("round-trips through parseDateRangeSearchParam", () => {
    const date = local(2026, 9, 1, 14, 30);
    expect(
      parseDateRangeSearchParam(serializeDateRangeSearchParam(date), "to")
    ).toEqual(date);
  });
});

describe("detectDateRangePreset", () => {
  const reference = local(2026, 9, 10, 15, 0);

  it("matches a preset only on the exact instants", () => {
    const { from, to } = getPresetDateRange("last7Days", reference);
    expect(detectDateRangePreset(from, to, reference)).toEqual("last7Days");
  });

  it("treats a same-day range with a different time as custom", () => {
    expect(
      detectDateRangePreset(
        local(2026, 9, 10, 0, 0),
        local(2026, 9, 10, 14, 0),
        reference
      )
    ).toEqual("custom");
  });
});

// The from/to query params were historically written with
// DateTime.toISODate(), i.e. local calendar dates. Reading them back must land
// on the same local calendar day, otherwise the range shifts by a day on every
// reload in UTC-negative zones.
describe("getInitialAuditLogDateRange", () => {
  it("round-trips a legacy custom range as whole inclusive local days", () => {
    const { preset, rangeFrom, rangeTo } = getInitialAuditLogDateRange(
      "2026-09-01",
      "2026-09-03",
      String(new Date("2026-09-10T12:00:00Z").getTime())
    );

    expect(preset).toEqual("custom");
    expect(rangeFrom).toEqual(local(2026, 9, 1, 0, 0));
    expect(rangeTo).toEqual(local(2026, 9, 3, 23, 59));
  });

  it("keeps the time of an ISO datetime custom range", () => {
    const from = local(2026, 9, 1, 9, 15);
    const to = local(2026, 9, 1, 17, 45);
    const { preset, rangeFrom, rangeTo } = getInitialAuditLogDateRange(
      serializeDateRangeSearchParam(from),
      serializeDateRangeSearchParam(to),
      String(local(2026, 9, 10, 12).getTime())
    );

    expect(preset).toEqual("custom");
    expect(rangeFrom).toEqual(from);
    expect(rangeTo).toEqual(to);
  });

  it("returns the today preset when both params are absent", () => {
    const { preset } = getInitialAuditLogDateRange(null, null, null);
    expect(preset).toEqual("today");
  });
});

describe("formatCustomDateRangeLabel", () => {
  it("shows dates only when the range covers whole days", () => {
    expect(
      formatCustomDateRangeLabel(
        "en",
        local(2026, 9, 1, 0, 0),
        local(2026, 9, 3, 23, 59)
      )
    ).toEqual("Sep 1 – Sep 3, 2026");
  });

  it("shows a single date when a whole-day range is one day", () => {
    expect(
      formatCustomDateRangeLabel(
        "en",
        local(2026, 9, 1, 0, 0),
        local(2026, 9, 1, 23, 59)
      )
    ).toEqual("Sep 1, 2026");
  });

  it("includes the times when the range is not whole days", () => {
    const label = formatCustomDateRangeLabel(
      "en",
      local(2026, 9, 1, 14, 30),
      local(2026, 9, 3, 23, 59)
    );
    expect(label).not.toBeNull();
    // Intl may separate the meridiem with a narrow no-break space.
    const normalized = label!.replace(/\s/g, " ");
    expect(normalized).toEqual("Sep 1, 2:30 PM – Sep 3, 2026, 11:59 PM");
  });

  it("returns null when either bound is missing", () => {
    expect(
      formatCustomDateRangeLabel("en", null, local(2026, 9, 3))
    ).toBeNull();
  });
});
