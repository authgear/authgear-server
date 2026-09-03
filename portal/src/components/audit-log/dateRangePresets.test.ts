import { describe, it, expect } from "@jest/globals";
import { DateTime } from "luxon";
import { getInitialAuditLogDateRange } from "./dateRangePresets";

// The from/to query params are written with DateTime.toISODate(), i.e. local
// calendar dates. Reading them back must land on the same local calendar day,
// otherwise the range shifts by a day on every reload in UTC-negative zones.
describe("getInitialAuditLogDateRange", () => {
  it("round-trips a custom range without shifting the local calendar day", () => {
    const { preset, rangeFrom, rangeTo } = getInitialAuditLogDateRange(
      "2026-09-01",
      "2026-09-03",
      String(new Date("2026-09-10T12:00:00Z").getTime())
    );

    expect(preset).toEqual("custom");
    expect(DateTime.fromJSDate(rangeFrom!).toISODate()).toEqual("2026-09-01");
    expect(DateTime.fromJSDate(rangeTo!).toISODate()).toEqual("2026-09-03");
  });

  it("returns the today preset when both params are absent", () => {
    const { preset } = getInitialAuditLogDateRange(null, null, null);
    expect(preset).toEqual("today");
  });
});
