// Client-side password generator mirrors server-side password generator in
// authgear-server/pkg/lib/authn/authenticator/password/generator.go

import { validatePassword } from "../error/password";
import { PasswordPolicyConfig } from "../types";

// Character list for each category.
const CharListLowercase = "abcdefghijklmnopqrstuvwxyz";
const CharListUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
const CharListAlphabet = CharListLowercase + CharListUppercase;
const CharListDigit = "0123456789";
const CharListAlphanumeric = CharListAlphabet + CharListDigit;
// Referenced from "special" character class in Apple's Password Autofill rules.
// https://developer.apple.com/documentation/security/password_autofill/customizing_password_autofill_rules
const CharListSymbol = "-~!@#$%^&*_+=`|(){}[:;\"'<>,.?]";

const DefaultCharList = CharListAlphanumeric;
const MaxTrials = 10;
const DefaultMinLength = 8;
const GuessableEnabledMinLength = 32;

export interface RandSource {
  pick(list: string): number;
  shuffle(list: string): string;
}

class CryptoRandSource implements RandSource {
  // eslint-disable-next-line class-methods-use-this
  private uint32(): number {
    const array = new Uint32Array(1);
    window.crypto.getRandomValues(array);
    return array[0];
  }

  // eslint-disable-next-line class-methods-use-this
  pick(list: string): number {
    const n = list.length;
    const discard = Math.pow(2, 32) - (Math.pow(2, 32) % n);
    let v = this.uint32();
    while (v >= discard) {
      v = this.uint32();
    }
    return v % n;
  }

  shuffle(list: string): string {
    const array = list.split("");
    for (let i = array.length - 1; i > 0; i--) {
      const j = this.pick(array.slice(0, i + 1).join(""));
      [array[i], array[j]] = [array[j], array[i]];
    }
    return array.join("");
  }
}

export class PasswordGenerator {
  passwordPolicy: PasswordPolicyConfig;
  randSource: RandSource;

  constructor(
    passwordPolicy: PasswordPolicyConfig,
    randSource: RandSource = new CryptoRandSource()
  ) {
    this.randSource = randSource;
    this.passwordPolicy = passwordPolicy;
  }

  generate(): string | null {
    for (let i = 0; i < MaxTrials; i++) {
      const password = this._generate();
      if (validatePassword(password, this.passwordPolicy) === null) {
        return password;
      }
    }
    return null;
  }

  private _generate(): string {
    const { passwordPolicy } = this;

    const charList = this.prepareCharList();
    const minLength = this._determineMinLength(passwordPolicy);

    let password = this._addRequiredCharacters(passwordPolicy);
    password = this._fillRemainingCharacters(password, charList, minLength);

    return this.randSource.shuffle(password);
  }

  // eslint-disable-next-line class-methods-use-this
  private _determineMinLength(passwordPolicy: PasswordPolicyConfig): number {
    if (
      passwordPolicy.minimum_guessable_level !== undefined &&
      passwordPolicy.minimum_guessable_level > 0
    ) {
      return GuessableEnabledMinLength;
    }
    return passwordPolicy.min_length !== undefined &&
      passwordPolicy.min_length > DefaultMinLength
      ? passwordPolicy.min_length
      : DefaultMinLength;
  }

  // eslint-disable-next-line class-methods-use-this
  private _addRequiredCharacters(passwordPolicy: PasswordPolicyConfig): string {
    let password = "";
    if (passwordPolicy.uppercase_required)
      password += CharListUppercase[this.randSource.pick(CharListUppercase)];
    if (passwordPolicy.lowercase_required)
      password += CharListLowercase[this.randSource.pick(CharListLowercase)];
    if (
      passwordPolicy.alphabet_required &&
      !passwordPolicy.uppercase_required &&
      !passwordPolicy.lowercase_required
    )
      password += CharListAlphabet[this.randSource.pick(CharListAlphabet)];
    if (passwordPolicy.digit_required)
      password += CharListDigit[this.randSource.pick(CharListDigit)];
    if (passwordPolicy.symbol_required)
      password += CharListSymbol[this.randSource.pick(CharListSymbol)];
    return password;
  }

  private _fillRemainingCharacters(
    password: string,
    charList: string,
    minLength: number
  ): string {
    for (let i = password.length; i < minLength; i++) {
      password += charList[this.randSource.pick(charList)];
    }
    return password;
  }

  private prepareCharList(): string {
    const { passwordPolicy } = this;

    let charList = DefaultCharList;
    if (passwordPolicy.lowercase_required) charList += CharListLowercase;
    if (passwordPolicy.uppercase_required) charList += CharListUppercase;
    if (passwordPolicy.alphabet_required) charList += CharListAlphabet;
    if (passwordPolicy.digit_required) charList += CharListDigit;
    if (passwordPolicy.symbol_required) charList += CharListSymbol;

    passwordPolicy.excluded_keywords?.forEach((keyword) => {
      charList = charList.replace(new RegExp(keyword, "g"), "");
    });

    return charList;
  }
}
