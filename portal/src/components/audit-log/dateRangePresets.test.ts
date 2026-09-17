import { describe, it, expect } from "@jest/globals";
import { DateTime } from "luxon";
import {
  detectDateRangePreset,
  formatCustomDateRangeLabel,
  formatDateRangeBoundLabel,
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

// Intl may separate the meridiem with a narrow no-break space.
function normalizeSpaces(s: string | null): string | null {
  return s == null ? null : s.replace(/\s/g, " ");
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

  it("tolerates a plus offset that was decoded to a space", () => {
    // A literal "+" in a hand-pasted URL is decoded as a space by
    // URLSearchParams; the offset is the only place that can happen.
    expect(
      parseDateRangeSearchParam("2026-09-01T14:30:00 08:00", "to")
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

  it("treats an open-ended range as custom", () => {
    expect(
      detectDateRangePreset(local(2026, 9, 10, 0, 0), null, reference)
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
        local(2026, 9, 3, 23, 59),
        true
      )
    ).toEqual("Sep 1 – Sep 3, 2026");
  });

  it("shows a single date when a whole-day range is one day", () => {
    expect(
      formatCustomDateRangeLabel(
        "en",
        local(2026, 9, 1, 0, 0),
        local(2026, 9, 1, 23, 59),
        true
      )
    ).toEqual("Sep 1, 2026");
  });

  it("includes the times when the range is not whole days", () => {
    expect(
      normalizeSpaces(
        formatCustomDateRangeLabel(
          "en",
          local(2026, 9, 1, 14, 30),
          local(2026, 9, 3, 23, 59),
          true
        )
      )
    ).toEqual("Sep 1, 2:30 PM – Sep 3, 2026, 11:59 PM");
  });

  it("stays date-only when times are not shown, even for timed bounds", () => {
    // Analytics stores UTC midnights, which are not local day boundaries.
    expect(
      formatCustomDateRangeLabel(
        "en",
        local(2026, 9, 1, 8, 0),
        local(2026, 9, 3, 8, 0),
        false
      )
    ).toEqual("Sep 1 – Sep 3, 2026");
    expect(
      formatCustomDateRangeLabel(
        "en",
        local(2026, 9, 1, 8, 0),
        local(2026, 9, 1, 8, 0),
        false
      )
    ).toEqual("Sep 1, 2026");
  });

  it("returns null when either bound is missing", () => {
    expect(
      formatCustomDateRangeLabel("en", null, local(2026, 9, 3), true)
    ).toBeNull();
  });
});

describe("formatDateRangeBoundLabel", () => {
  it("shows the date only for a start bound at the start of its day", () => {
    expect(
      formatDateRangeBoundLabel("en", local(2026, 9, 8, 0, 0), "from", true)
    ).toEqual("Sep 8, 2026");
  });

  it("shows the date only for an end bound at the end of its day", () => {
    expect(
      formatDateRangeBoundLabel("en", local(2026, 9, 9, 23, 59), "to", true)
    ).toEqual("Sep 9, 2026");
  });

  it("includes the time for a bound that is not on its day boundary", () => {
    expect(
      normalizeSpaces(
        formatDateRangeBoundLabel("en", local(2026, 9, 8, 22, 0), "from", true)
      )
    ).toEqual("Sep 8, 2026, 10:00 PM");
  });

  it("treats a start bound at 23:59 as timed, not whole-day", () => {
    expect(
      normalizeSpaces(
        formatDateRangeBoundLabel("en", local(2026, 9, 8, 23, 59), "from", true)
      )
    ).toEqual("Sep 8, 2026, 11:59 PM");
  });

  it("stays date-only when times are not shown", () => {
    expect(
      formatDateRangeBoundLabel("en", local(2026, 9, 8, 22, 0), "from", false)
    ).toEqual("Sep 8, 2026");
  });
});
