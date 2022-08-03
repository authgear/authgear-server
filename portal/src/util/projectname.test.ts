import { describe, it, expect } from "@jest/globals";
import {
  determineWord,
  getRandom32BitsNumber,
  maskNumber,
  randomProjectName,
} from "./projectname";

// eslint-disable-next-line no-undef
const mGetRandomValues = jest.fn().mockReturnValueOnce(new Uint32Array(1));
Object.defineProperty(window, "crypto", {
  value: { getRandomValues: mGetRandomValues },
});

describe("determineWord", () => {
  it("handle valid index", () => {
    expect(determineWord(10)).toEqual("access");
  });

  it("handle valid index", () => {
    expect(determineWord(10000)).toEqual(undefined);
  });
});

describe("getRandom32BitsNumber", () => {
  it("generate random 32 bits number", () => {
    const num = getRandom32BitsNumber();
    expect(num).toBeGreaterThanOrEqual(0);
    expect(num).toBeLessThanOrEqual(Math.pow(2, 32));
  });
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
