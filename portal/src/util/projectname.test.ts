import { describe, it, expect } from "@jest/globals";
import { maskNumber, deterministicProjectName } from "./projectname";

describe("maskNumber", () => {
  it("should mask the number starting from 0 bit and get 11 bits after it", () => {
    expect(maskNumber(12345678, 0, 11)).toEqual(334);
  });

  it("should mask the number starting from 20 bit and get 5 bits after it", () => {
    expect(maskNumber(12345678, 20, 5)).toEqual(11);
  });
});

describe("deterministicProjectName", () => {
  it("deterministicProjectName(0) should starts with 'abandon-' and ends with 6 lowercase-alphanumeric characters", () => {
    // 0 is 0b00000000000_00000000000_0000000000
    // 0b00000000000 is abandon
    // So the name starts with abandon
    expect(deterministicProjectName(0)).toMatch(/^abandon-[0-9a-z]{6}$/);
  });

  it("deterministicProjectName(0) should starts with 'abandon-' and ends with 6 lowercase-alphanumeric characters", () => {
    // 1 is 0b00000000000_00000000000_0000000001
    // 0b00000000000 is abandon
    // So the name starts with abandon
    expect(deterministicProjectName(1)).toMatch(/^abandon-[0-9a-z]{6}$/);
  });

  it("deterministicProjectName(87878787) should starts with 'ahead' and ends with 6 lowercase-alphanumeric characters", () => {
    // 87878787 is 0b00000101001_11100111011_0010000011
    // 0b00000101001 is ahead
    // So the name starts with ahead
    expect(deterministicProjectName(87878787)).toMatch(/^ahead-[0-9a-z]{6}$/);
  });

  it("deterministicProjectName(4294967295) should starts with 'zoo-' and ends with 6 lowercase-alphanumeric characters", () => {
    // 4294967295 (2^32 - 1) is 0b11111111111_11111111111_1111111111
    // 0b11111111111 is zoo
    // So the name starts with zoo-
    expect(deterministicProjectName(4294967295)).toMatch(/^zoo-[0-9a-z]{6}$/);
  });
});
