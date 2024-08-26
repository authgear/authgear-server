import { describe, it, expect } from "@jest/globals";
import {
  deterministicProjectName,
  extractBits,
  projectNameWithCompanyName,
} from "./projectname";

describe("extractBits", () => {
  it("should handle bytes that look like signed bits", () => {
    expect(
      extractBits(
        new Uint8Array([
          0b10000000, 0b10000001, 0b10000010, 0b10000011, 0b10000100,
          0b10000101,
        ])
      )
    ).toEqual([0b10000000100, 0b0000110000010100000111000010010]);
  });

  it("should handle bytes", () => {
    expect(
      extractBits(
        new Uint8Array([
          0b00000000, 0b00000001, 0b00000010, 0b00000011, 0b00000100,
          0b00000101,
        ])
      )
    ).toEqual([0b00000000000, 0b0000100000010000000110000010000]);
  });
});

describe("deterministicProjectName", () => {
  it("deterministicProjectName([0, 0, 0, 0, 0, 0]) is 'abandon-000000'", () => {
    // [0, 0, 0, 0, 0, 0] is 0b00000000000_0000000000000000000000000000000_000000
    // 0b00000000000 is 'abandon'
    // 0000000000000000000000000000000 is '000000'
    // last 6 bits 000000 is not used
    // So the name is 'abandon-000000'
    const fortyEightBits = new Uint8Array([0, 0, 0, 0, 0, 0]);
    expect(deterministicProjectName(fortyEightBits)).toEqual("abandon-000000");
  });

  it("deterministicProjectName([0, 0, 0, 0, 1, 0]) is 'abandon-000004'", () => {
    // [0, 0, 0, 0, 1, 0] is 0b00000000000_0000000000000000000000000000100_000000
    // 0b00000000000 is 'abandon'
    // 0000000000000000000000000000100 is '000004'
    // last 6 bits 000000 is not used
    // So the name is 'abandon-000000'
    const fortyEightBits = new Uint8Array([0, 0, 0, 0, 1, 0]);
    expect(deterministicProjectName(fortyEightBits)).toEqual("abandon-000004");
  });

  it("deterministicProjectName([87, 87, 87, 87, 87, 87]) is 'firm-pwldul' ", () => {
    // [87, 87, 87, 87, 87, 87] is 0b01010111010_1011101010111010101110101011101_010111
    // 0b0b01010111010 is 'firm'
    // 1011101010111010101110101011101 is 'pwldul'
    // last 6 bits 010111 is not used
    // So the name is 'firm-000000'
    const fortyEightBits = new Uint8Array([87, 87, 87, 87, 87, 87]);
    expect(deterministicProjectName(fortyEightBits)).toEqual("firm-pwldul");
  });
});

describe("projectNameWithCompanyName", () => {
  it("projectNameWithCompanyName('authgear') starts with 'authgear-` and ends with 6 lowercase-alphanumeric characters", () => {
    const authgearProjectName = projectNameWithCompanyName("authgear");
    expect(authgearProjectName).toMatch(/^authgear-[a-z0-9]{6}$/);
  });
});
