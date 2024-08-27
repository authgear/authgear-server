import { parseEmail } from "./email";

/* global it, expect */
it("parseEmail", () => {
  // invalid
  expect(parseEmail("")).toEqual(null); // need @
  expect(parseEmail("@")).toEqual(null); // need chars before & after @
  expect(parseEmail("a@")).toEqual(null); // need chars after @
  expect(parseEmail("@b")).toEqual(null); // need chars before @
  expect(parseEmail("a @ b")).toEqual(null); // no space in middle

  // valid
  expect(parseEmail("a@b")).toEqual("a@b"); // valid
  expect(parseEmail("   a@b     ")).toEqual("a@b"); // trimmed
  expect(parseEmail("\t\r\n+authgear@authgear.com\t\r\n")).toEqual(
    "+authgear@authgear.com"
  ); // trimmed
});
