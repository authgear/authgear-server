import { describe, it, expect } from "@jest/globals";
import { maskNumber, randomProjectName } from "./projectname";

// eslint-disable-next-line no-undef
const mGetRandomValues = jest.fn().mockReturnValueOnce(new Uint32Array(1));
Object.defineProperty(window, "crypto", {
  value: { getRandomValues: mGetRandomValues },
});

describe("maskNumber", () => {
  it("should mask the number starting from 0 bit and get 11 bits after it", () => {
    expect(maskNumber(12345678, 0, 11)).toEqual(334);
  });

  it("should mask the number starting from 20 bit and get 5 bits after it", () => {
    expect(maskNumber(12345678, 20, 5)).toEqual(11);
  });
});

describe("randomProjectName", () => {
  it("should generate random project name", () => {
    expect(randomProjectName()).toEqual("abandon-abandon-0");
  });
});
