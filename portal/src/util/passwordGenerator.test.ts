import { describe, it, expect } from "@jest/globals";
import { PasswordGenerator, RandSource } from "./passwordGenerator";
import * as zxcvbn from "zxcvbn";
import { extractGuessableLevel } from "../PasswordField";

describe("passwordGenerator", () => {
  const fixedRandSource: RandSource = {
    pick: (_: string) => 0,
    shuffle: (list: string) => list,
  };

  it("falls back to default settings", () => {
    const password = new PasswordGenerator({}, fixedRandSource).generate();
    expect(password).not.toBeNull();
    expect(password!.length).toBe(8);
  });

  it("respects uppercase requirement", () => {
    const password = new PasswordGenerator(
      { uppercase_required: true },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).toMatch(/[A-Z]/);
  });

  it("respects lowercase requirement", () => {
    const password = new PasswordGenerator(
      { lowercase_required: true },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).toMatch(/[a-z]/);
  });

  it("respects number requirement", () => {
    const password = new PasswordGenerator(
      { digit_required: true },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).toMatch(/[0-9]/);
  });

  it("respects special character requirement", () => {
    const password = new PasswordGenerator(
      { symbol_required: true },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).toMatch(/[^A-Za-z0-9]/);
  });

  it("respects length requirement", () => {
    const password = new PasswordGenerator(
      { min_length: 40 },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password!.length).toBe(40);
  });

  it("respects combined requirements", () => {
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

  it("respects minimum guessable level", () => {
    const password = new PasswordGenerator({
      minimum_guessable_level: 3,
    }).generate();
    expect(password).not.toBeNull();
    const result = zxcvbn(password!, []);
    const guessableLevel = extractGuessableLevel(result);
    expect(guessableLevel).toBeGreaterThanOrEqual(3);
  });

  it("respects exclusion list", () => {
    const password = new PasswordGenerator(
      {
        digit_required: true,
        excluded_keywords: ["1", "2", "3", "4", "5", "6", "7", "8", "9"],
      },
      fixedRandSource
    ).generate();
    expect(password).not.toBeNull();
    expect(password).not.toMatch(/[123456789]/);
    expect(password).toMatch(/0/);
  });
});
