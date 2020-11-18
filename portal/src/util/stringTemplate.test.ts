/* global describe, it, expect */
import { resourcePath } from "./stringTemplate";

describe("render and parse template string", () => {
  it("round trip", () => {
    const template = resourcePath`templates/${"locale"}/${"type"}/dummy.html`;
    const input = "templates/en/messages/dummy.html";
    const expectedResult = {
      locale: "en",
      type: "messages",
    };
    const parsed = template.parse(input);
    expect(parsed).toEqual(expectedResult);

    const rendered = template.render(parsed!);
    expect(rendered).toEqual(input);
  });
});
