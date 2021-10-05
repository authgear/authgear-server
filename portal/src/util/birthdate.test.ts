/* global describe, it, expect */
import { parseBirthdate, toBirthdate } from "./birthdate";

describe("parseBirthdate", () => {
  it("parses yyyy-MM-dd", () => {
    expect(parseBirthdate("1990-01-01")).toEqual(
      new Date("1990-01-01T00:00:00Z")
    );
  });

  it("does not parses 0000-MM-dd", () => {
    expect(parseBirthdate("0000-01-01")).toEqual(undefined);
  });

  it("does not parses --MM-dd", () => {
    expect(parseBirthdate("--01-01")).toEqual(undefined);
  });
});

describe("toBirthdate", () => {
  it("handles valid date", () => {
    expect(toBirthdate(new Date("1990-01-01T01:02:03Z"))).toEqual("1990-01-01");
  });

  it("handles invalid date", () => {
    const d = new Date();
    d.setTime(NaN);
    expect(toBirthdate(d)).toEqual(undefined);
  });
});
