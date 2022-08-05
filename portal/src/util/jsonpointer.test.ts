import { describe, it, expect } from "@jest/globals";
import {
  parseJSONPointer,
  jsonPointerToString,
  parseJSONPointerIntoParentChild,
  parentChildToJSONPointer,
  matchParentChild,
  joinParentChild,
} from "./jsonpointer";

describe("parseJSONPointer", () => {
  it("parse", () => {
    const f = parseJSONPointer;
    expect(f("")).toEqual([]);
    expect(f("/")).toEqual([""]);
    expect(f("//")).toEqual(["", ""]);
    expect(f("/a")).toEqual(["a"]);
    expect(f("/a~0")).toEqual(["a~"]);
    expect(f("/a~1")).toEqual(["a/"]);
    expect(f("/a~1b")).toEqual(["a/b"]);
  });
});

describe("jsonPointerToString", () => {
  it("stringify", () => {
    const f = jsonPointerToString;
    expect(f([])).toEqual("");
    expect(f([""])).toEqual("/");
    expect(f(["", ""])).toEqual("//");
    expect(f(["a"])).toEqual("/a");
    expect(f(["a~"])).toEqual("/a~0");
    expect(f(["a/"])).toEqual("/a~1");
    expect(f(["a/b"])).toEqual("/a~1b");
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

describe("parentChildToJSONPointer", () => {
  it("work", () => {
    const f = parentChildToJSONPointer;
    expect(f("", "")).toEqual("/");
    expect(f("", "a")).toEqual("/a");
    expect(f("/a", "b")).toEqual("/a/b");
  });
});

describe("matchParentChild", () => {
  it("match", () => {
    const f = matchParentChild;
    expect(f("/a", "", "a")).toEqual(true);
    expect(f("", "", "a")).toEqual(false);
    expect(f("/a/b", "/a", "b")).toEqual(true);
    expect(f("/a/b", /^\/a$/, "b")).toEqual(true);
    expect(f("/secrets/2/data/a", /^\/secrets\/\d+\/data$/, "a")).toEqual(true);
  });
});

describe("joinParentChild", () => {
  it("join", () => {
    const f = joinParentChild;
    expect(f("", "a")).toEqual("/a");
    expect(f(/\/a/, "b")).toEqual(/\/a\/b/);
    expect(f(/\/a\/\d+/, "b")).toEqual(/\/a\/\d+\/b/);
  });
});
