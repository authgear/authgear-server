/* global describe, it, expect */
import { isoWeekLabel, monthLabel } from "./date";

describe("isoWeekLabel", () => {
  it("convert ios to ios week label YYYY-Www", () => {
    expect(isoWeekLabel("1990-01-01")).toEqual("1990-01-01 (W01)");
    expect(isoWeekLabel("2000-01-01")).toEqual("2000-01-01 (W52)");
  });
});

describe("monthLabel", () => {
  it("convert monthLabel", () => {
    expect(monthLabel("1990-01-01")).toEqual("Jan 1990");
  });
});
