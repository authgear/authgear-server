/* global describe, it, expect */
import { exactKeywordSearch } from "./search";

describe("exactKeywordSearch", () => {
  it("does not output duplicate items", () => {
    const actual = exactKeywordSearch(
      [
        {
          nameA: "foobar",
          nameB: "foobar",
        },
        {
          nameA: "baz",
          nameB: "42",
        },
      ],
      ["nameA", "nameB"],
      "foobar"
    );
    const expected = [
      {
        nameA: "foobar",
        nameB: "foobar",
      },
    ];
    expect(actual).toEqual(expected);
  });
});
