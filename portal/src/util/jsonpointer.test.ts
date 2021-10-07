/* global describe, it, expect */
import {
  parseJSONPointer,
  jsonPointerToString,
  parseJSONPointerIntoParentChild,
} from "./jsonpointer";

describe("parseJSONPointer", () => {
  it("parse", () => {
    const f = parseJSONPointer;
    expect(f("")).toEqual([]);
    expect(f("/")).toEqual([""]);
    expect(f("//")).toEqual(["", ""]);
    expect(f("/a")).toEqual(["a"]);
  });
});

describe("jsonPointerToString", () => {
  it("stringify", () => {
    const f = jsonPointerToString;
    expect(f([])).toEqual("");
    expect(f([""])).toEqual("/");
    expect(f(["", ""])).toEqual("//");
    expect(f(["a"])).toEqual("/a");
  });
});

describe("parseJSONPointerIntoParentChild", () => {
  it("work", () => {
    const f = parseJSONPointerIntoParentChild;
    expect(f("")).toEqual(null);
    expect(f("/")).toEqual(["", ""]);
    expect(f("/a")).toEqual(["", "a"]);
    expect(f("/a/b")).toEqual(["/a", "b"]);
  });
});
