import { validateEmail } from "./email";

/* global it, expect */
it("validateEmail", () => {
  // invalid
  expect(validateEmail("")).toEqual(null); // need @
  expect(validateEmail("@")).toEqual(null); // need chars before & after @
  expect(validateEmail("a@")).toEqual(null); // need chars after @
  expect(validateEmail("@b")).toEqual(null); // need chars before @
  expect(validateEmail("a @ b")).toEqual(null); // no space in middle

  // valid
  expect(validateEmail("a@b")).toEqual("a@b"); // valid
  expect(validateEmail("   a@b     ")).toEqual("a@b"); // trimmed
  expect(validateEmail("\t\r\n+authgear@authgear.com\t\r\n")).toEqual(
    "+authgear@authgear.com"
  ); // trimmed
});
