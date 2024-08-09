import { describe, it, expect } from "@jest/globals";
import { determineMinLength, PasswordGenerator, prepareCharList, RandSource } from "./passwordGenerator";
import { PasswordPolicyConfig } from '../types';
import { zxcvbnGuessableLevel } from './zxcvbn';

describe("passwordGenerator", () => {
  const fixedRandSource: RandSource = {
    randomBytes: (n: number) => new Uint8Array(n),
    shuffle: (list: string) => list,
  };

  it("should generate a password with default settings", () => {
    const password = new PasswordGenerator({}, fixedRandSource).generate();
    expect(password).not.toBeNull();
    expect(password!.length).toBeGreaterThanOrEqual(8);
  });

  it("should include at least one uppercase letter when required", () => {
    const password = new PasswordGenerator(
      { uppercase_required: true },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).toMatch(/[A-Z]/);
  });

  it("should include at least one lowercase letter when required", () => {
    const password = new PasswordGenerator(
      { lowercase_required: true },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).toMatch(/[a-z]/);
  });

  it("should include at least one digit when required", () => {
    const password = new PasswordGenerator(
      { digit_required: true },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).toMatch(/[0-9]/);
  });

  it("should include at least one special character when required", () => {
    const password = new PasswordGenerator(
      { symbol_required: true },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).toMatch(/[^A-Za-z0-9]/);
  });

  it("should meet the minimum length requirement", () => {
    const password = new PasswordGenerator(
      { min_length: 40 },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password!.length).toBeGreaterThanOrEqual(40);
  });

  it("should meet all combined requirements", () => {
    const password = new PasswordGenerator(
      {
        uppercase_required: true,
        lowercase_required: true,
        digit_required: true,
        symbol_required: true,
        min_length: 12,
      },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).toMatch(/[A-Z]/);
    expect(password).toMatch(/[a-z]/);
    expect(password).toMatch(/[0-9]/);
    expect(password).toMatch(/[^A-Za-z0-9]/);
    expect(password!.length).toBe(12);
  });

  it("should meet the minimum guessable level requirement", () => {
    const password = new PasswordGenerator({
      minimum_guessable_level: 4,
    }).generate();
    expect(password).not.toBeNull();
    const guessableLevel = zxcvbnGuessableLevel(password!);
    expect(password!.length).toBeGreaterThanOrEqual(32);
    expect(guessableLevel).toBeGreaterThanOrEqual(4);
  });
});

describe("prepareCharList", () => {
  it("should return alphanumeric characters when no specific requirements are set", () => {
    const policy: PasswordPolicyConfig = {};
    const result = prepareCharList(policy);
    expect(result).toEqual("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789");
  });

  it("should return lowercase characters when lowercase is required", () => {
    const policy: PasswordPolicyConfig = { lowercase_required: true };
    const result = prepareCharList(policy);
    expect(result).toEqual("abcdefghijklmnopqrstuvwxyz");
  });

  it("should return uppercase characters when uppercase is required", () => {
    const policy: PasswordPolicyConfig = { uppercase_required: true };
    const result = prepareCharList(policy);
    expect(result).toEqual("ABCDEFGHIJKLMNOPQRSTUVWXYZ");
  });

  it("should return alphabet characters when alphabet is required", () => {
    const policy: PasswordPolicyConfig = { alphabet_required: true };
    const result = prepareCharList(policy);
    expect(result).toEqual("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ");
  });

  it("should return alphabet and lowercase characters when both are required", () => {
    const policy: PasswordPolicyConfig = { alphabet_required: true, lowercase_required: true };
    const result = prepareCharList(policy);
    expect(result).toEqual("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ");
  });

  it("should return alphabet and uppercase characters when both are required", () => {
    const policy: PasswordPolicyConfig = { alphabet_required: true, uppercase_required: true };
    const result = prepareCharList(policy);
    expect(result).toEqual("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ");
  });

  it("should return digit characters when digits are required", () => {
    const policy: PasswordPolicyConfig = { digit_required: true };
    const result = prepareCharList(policy);
    expect(result).toEqual("0123456789");
  });

  it("should return symbol characters when symbols are required", () => {
    const policy: PasswordPolicyConfig = { symbol_required: true };
    const result = prepareCharList(policy);
    expect(result).toEqual("-~!@#$%^&*_+=`|(){}[:;\"'<>,.?]");
  });

  it("should return all character sets when all are required", () => {
    const policy: PasswordPolicyConfig = {
      lowercase_required: true,
      uppercase_required: true,
      alphabet_required: true,
      digit_required: true,
      symbol_required: true,
    };
    const result = prepareCharList(policy);
    expect(result).toEqual("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-~!@#$%^&*_+=`|(){}[:;\"'<>,.?]");
  });
});

describe("determineMinLength", () => {
  it("should return minLength when it is greater than DefaultMinLength and GuessableEnabledMinLength", () => {
    const policy: PasswordPolicyConfig = {
      min_length: 15,
      minimum_guessable_level: 0,
    };
    const result = determineMinLength(policy);
    expect(result).toEqual(15);
  });

  it("should return DefaultMinLength when minLength is less than DefaultMinLength", () => {
    const policy: PasswordPolicyConfig = {
      min_length: 5,
      minimum_guessable_level: 0,
    };
    const result = determineMinLength(policy);
    expect(result).toEqual(8);
  });

  it("should return GuessableEnabledMinLength when minLength is less than GuessableEnabledMinLength and minimum_guessable_level is greater than 0", () => {
    const policy: PasswordPolicyConfig = {
      min_length: 10,
      minimum_guessable_level: 1,
    };
    const result = determineMinLength(policy);
    expect(result).toEqual(32);
  });
});
