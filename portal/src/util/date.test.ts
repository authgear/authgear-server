import { describe, it, expect } from "@jest/globals";
import { DateTime } from "luxon";
import {
  isoWeekLabels,
  monthLabel,
  parseDate,
  utcDateToLocalDate,
} from "./date";

describe("isoWeekLabels", () => {
  it("convert ios to ios week label YYYY-Www", () => {
    expect(isoWeekLabels("1990-01-01")).toEqual(["1990-01-01", "(W01)"]);
    expect(isoWeekLabels("2000-01-01")).toEqual(["2000-01-01", "(W52)"]);
  });
});

describe("monthLabel", () => {
  it("convert monthLabel", () => {
    expect(monthLabel("1990-01-01")).toEqual("Jan 1990");
  });
});

describe("parseDate", () => {
  it("parse iso date string to js date", () => {
    expect(parseDate("1990-01-01")).toEqual(
      new Date(Date.UTC(1990, 0, 1, 0, 0, 0))
    );
  });
});

describe("utcDateToLocalDate", () => {
  it("keeps the UTC calendar day in the local time zone", () => {
    for (const day of ["2026-01-01", "2026-03-08", "2026-11-24"]) {
      const local = utcDateToLocalDate(parseDate(day));
      expect(DateTime.fromJSDate(local).toISODate()).toEqual(day);
    }
  });
});
