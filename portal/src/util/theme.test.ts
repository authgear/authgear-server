/* global describe, it, expect */
import { getShades } from "./theme";

describe("getShades", () => {
  it("gives the same result as https://fabricweb.z5.web.core.windows.net/pr-deploy-site/refs/heads/master/theming-designer/index.html does", () => {
    const expected = [
      "#0078d4",
      "#f3f9fd",
      "#d0e7f8",
      "#a9d3f2",
      "#5ca9e5",
      "#1a86d9",
      "#006cbe",
      "#005ba1",
      "#004377",
    ];
    const actual = getShades("#0078d4");
    expect(actual).toEqual(expected);
  });
});
