import { validatePassword } from "../error/password";
import { PasswordPolicyConfig } from "../types";

// Character list for each category.
const CharListLowercase = "abcdefghijklmnopqrstuvwxyz";
const CharListUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
const CharListAlphabet = CharListLowercase + CharListUppercase;
const CharListDigit = "0123456789";
// Referenced from "special" character class in Apple's Password Autofill rules.
// https://developer.apple.com/documentation/security/password_autofill/customizing_password_autofill_rules
const CharListSymbol = "-~!@#$%^&*_+=`|(){}[:;\"'<>,.?]";

const MaxTrials = 10;
const DefaultMinLength = 8;
const GuessableEnabledMinLength = 32;

export interface RandSource {
  randomBytes(n: number): Uint8Array;
  shuffle(list: string): string;
}

class CryptoRandSource implements RandSource {
  randomBytes(n: number): Uint8Array {
    const array = new Uint8Array(n);
    window.crypto.getRandomValues(array);
    return array;
  }

  shuffle(list: string): string {
    const array = list.split("");
    for (let i = array.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
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

    const minLength = determineMinLength(passwordPolicy);
    const charList = prepareCharList(passwordPolicy);

    let password = this._addRequiredCharacters(passwordPolicy);
    password = this._fillRemainingCharacters(password, charList, minLength);

    return this.randSource.shuffle(password);
  }

  private _addRequiredCharacters(passwordPolicy: PasswordPolicyConfig): string {
    let password = "";
    if (passwordPolicy.lowercase_required) {
      password += this.pickRandChar(CharListLowercase);
    }
    if (passwordPolicy.uppercase_required) {
      password += this.pickRandChar(CharListUppercase);
    }
    if (passwordPolicy.alphabet_required && !passwordPolicy.lowercase_required && !passwordPolicy.uppercase_required) {
      password += this.pickRandChar(CharListAlphabet);
    }
    if (passwordPolicy.digit_required) {
      password += this.pickRandChar(CharListDigit);
    }
    if (passwordPolicy.symbol_required) {
      password += this.pickRandChar(CharListSymbol);
    }
    return password;
  }

  private _fillRemainingCharacters(password: string, charList: string, minLength: number): string {
    for (let i = password.length; i < minLength; i++) {
      password += this.pickRandChar(charList);
    }
    return password;
  }

  private pickRandChar(charList: string): string {
    const randomBytes = this.randSource.randomBytes(1);
    const maxByte = 256;
    const discard = maxByte - (maxByte % charList.length);
    let byte = randomBytes[0];

    while (byte >= discard) {
      byte = this.randSource.randomBytes(1)[0];
    }

    return charList[byte % charList.length];
  }
}

export function prepareCharList(passwordPolicy: PasswordPolicyConfig): string {
  const set = new Set<string>();

  if (passwordPolicy.alphabet_required) set.add(CharListAlphabet);
  if (passwordPolicy.lowercase_required) set.add(CharListLowercase);
  if (passwordPolicy.uppercase_required) set.add(CharListUppercase);
  if (passwordPolicy.digit_required) set.add(CharListDigit);
  if (passwordPolicy.symbol_required) set.add(CharListSymbol);

  // Default to alphanumeric if no character set is required.
  if (set.size === 0) {
    set.add(CharListAlphabet);
    set.add(CharListDigit);
  }

  // Remove overlapping character sets.
  if (set.has(CharListAlphabet)) {
    set.delete(CharListLowercase);
    set.delete(CharListUppercase);
  }

  // Build the final character list.
  let charList = '';
  set.forEach(cs => charList += cs);

  return charList;
}

export function determineMinLength(passwordPolicy: PasswordPolicyConfig): number {
  let minLength = passwordPolicy.min_length ?? DefaultMinLength;

  if (minLength < DefaultMinLength) {
    minLength = DefaultMinLength;
  }

  if ((passwordPolicy.minimum_guessable_level ?? 0) > 0 && minLength < GuessableEnabledMinLength) {
    minLength = GuessableEnabledMinLength;
  }

  return minLength;
}
