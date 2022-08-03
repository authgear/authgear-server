import { describe, it, expect } from "@jest/globals";
import {
  maskNumber,
  deterministicProjectName,
  randomProjectName,
} from "./projectname";

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

describe("deterministicProjectName", () => {
  it("deterministicProjectName(0) is 'abandon-abandon-0'", () => {
    // 0 is 0b00000000000_00000000000_0000000000
    // 0b00000000000 is abandon
    // 0b0000000000 is 0
    // So the name is abandon-abandon-0
    expect(deterministicProjectName(0)).toEqual("abandon-abandon-0");
  });

  it("deterministicProjectName(1) is ''", () => {
    // 1 is 0b00000000000_00000000000_0000000001
    // 0b00000000000 is abandon
    // 0b0000000001 is 1
    // So the name is abandon-abandon-1
    expect(deterministicProjectName(1)).toEqual("abandon-abandon-1");
  });

  it("deterministicProjectName(87878787) is ''", () => {
    // 87878787 is 0b00000101001_11100111011_0010000011
    // 0b00000101001 is ahead
    // 0b11100111011 is trash
    // 0b0010000011 is 131
    // So the name is ahead-trash-131
    expect(deterministicProjectName(87878787)).toEqual("ahead-trash-131");
  });

  it("deterministicProjectName(4294967295) is ''", () => {
    // 4294967295 (2^32 - 1) is 0b11111111111_11111111111_1111111111
    // 0b11111111111 is zoo
    // 0b1111111111 is 1023
    // So the name is zoo-zoo-1023
    expect(deterministicProjectName(4294967295)).toEqual("zoo-zoo-1023");
  });
});

describe("randomProjectName", () => {
  it("should generate random project name", () => {
    expect(randomProjectName()).toEqual("abandon-abandon-0");
  });
});
