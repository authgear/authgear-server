import { describe, it, expect } from "@jest/globals";
import { deduplicate } from "./array";

describe("deduplicate", () => {
  it("deduplicate", () => {
    expect(deduplicate(["1", "2", "3", "4"])).toEqual(["1", "2", "3", "4"]);
    expect(deduplicate(["1", "3", "2", "3", "4", "3"])).toEqual([
      "1",
      "3",
      "2",
      "4",
    ]);
    expect(deduplicate([1, 2, 1, 3, 4, 5, 2, 3])).toEqual([1, 2, 3, 4, 5]);
  });
});
