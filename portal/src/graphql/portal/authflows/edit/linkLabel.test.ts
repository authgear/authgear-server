import { describe, it, expect } from "@jest/globals";
import {
  isAllowedLinkURL,
  parseLinkLabel,
  serializeLinkLabel,
} from "./linkLabel";

describe("parseLinkLabel", () => {
  it("round-trips text with links", () => {
    const html = 'I agree to the <a href="https://e.com/t">Terms</a>.';
    const segments = parseLinkLabel(html);
    expect(segments).toEqual([
      { text: "I agree to the " },
      { text: "Terms", href: "https://e.com/t" },
      { text: "." },
    ]);
    expect(serializeLinkLabel(segments!)).toEqual(html);
  });

  it("refuses a link with other attributes", () => {
    expect(parseLinkLabel('<a href="x" onclick="y">T</a>')).toBeNull();
  });

  it("refuses other markup", () => {
    expect(parseLinkLabel("<b>x</b>")).toBeNull();
  });

  it("reads entities as text", () => {
    expect(parseLinkLabel("a &lt; b")).toEqual([{ text: "a < b" }]);
  });
});

describe("serializeLinkLabel", () => {
  it("escapes text", () => {
    expect(serializeLinkLabel([{ text: "a < b & c" }])).toEqual(
      "a &lt; b &amp; c"
    );
  });

  it("merges adjacent text and drops empty segments", () => {
    expect(
      serializeLinkLabel([
        { text: "a" },
        { text: "" },
        { text: "b" },
        { text: "", href: "https://e.com" },
      ])
    ).toEqual("ab");
  });

  it("escapes quotes in URLs", () => {
    expect(
      serializeLinkLabel([{ text: "T", href: 'https://e.com/?q="x"' }])
    ).toEqual('<a href="https://e.com/?q=&quot;x&quot;">T</a>');
  });
});

describe("isAllowedLinkURL", () => {
  it("accepts absolute http and https URLs only", () => {
    expect(isAllowedLinkURL("https://e.com")).toBe(true);
    expect(isAllowedLinkURL("http://e.com/p")).toBe(true);
    // Built from parts so the linter does not read it as a script URL.
    expect(isAllowedLinkURL(["javascript", "alert(1)"].join(":"))).toBe(false);
    expect(isAllowedLinkURL("/terms")).toBe(false);
    expect(isAllowedLinkURL("")).toBe(false);
  });
});

describe("parseLinkLabel and link URLs", () => {
  it("refuses a stored link that is not http(s)", () => {
    const script = ["javascript", "x()"].join(":");
    expect(parseLinkLabel(`<a href="${script}">T</a>`)).toBeNull();
    expect(parseLinkLabel('<a href="/terms">T</a>')).toBeNull();
  });
});
