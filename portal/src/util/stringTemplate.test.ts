/* global describe, it, expect */
import { renderTemplateString, parseTemplateString } from "./stringTemplate";

describe("render and parse template string", () => {
  it("round trip", () => {
    const template = "templates/{{ locale }}/{{ type }}/dummy.html";
    const input = "templates/en/messages/dummy.html";
    const expectedResult = {
      locale: "en",
      type: "messages",
    };
    const parsed = parseTemplateString(input, template);
    expect(parsed).toEqual(expectedResult);

    const rendered = renderTemplateString(
      parsed as Record<string, string>,
      template
    );
    expect(rendered).toEqual(input);
  });
});
