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

const cryptoRandSource: RandSource = {
  randomBytes: (n: number): Uint8Array => {
    const array = new Uint8Array(n);
    window.crypto.getRandomValues(array);
    return array;
  },
  shuffle: (list: string): string => {
    const array = list.split("");
    for (let i = array.length - 1; i > 0; i--) {
      const randomArray = new Uint8Array(1);
      window.crypto.getRandomValues(randomArray);
      const j = randomArray[0] % (i + 1);
      [array[i], array[j]] = [array[j], array[i]];
    }
    return array.join("");
  },
};

const mathRandSource: RandSource = {
  randomBytes: (n: number): Uint8Array => {
    const array = new Uint8Array(n);
    for (let i = 0; i < n; i++) {
      array[i] = Math.floor(Math.random() * 256);
    }
    return array;
  },
  shuffle: (list: string): string => {
    const array = list.split("");
    for (let i = array.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [array[i], array[j]] = [array[j], array[i]];
    }
    return array.join("");
  },
};

export function generatePassword(policy: PasswordPolicyConfig): string | null {
  try {
    window.crypto.getRandomValues(new Uint8Array(1));
    return generatePasswordWithSource(cryptoRandSource, policy);
  } catch {
    return generatePasswordWithSource(mathRandSource, policy);
  }
}

// eslint-disable-next-line complexity
export function generatePasswordWithSource(
  source: RandSource,
  policy: PasswordPolicyConfig
): string | null {
  for (let i = 0; i < MaxTrials; i++) {
    const minLength = determineMinLength(policy);
    const charList = prepareCharList(policy);

    let password = "";

    if (policy.lowercase_required) {
      password += pickRandChar(source, CharListLowercase);
    }
    if (policy.uppercase_required) {
      password += pickRandChar(source, CharListUppercase);
    }
    if (
      policy.alphabet_required &&
      !policy.lowercase_required &&
      !policy.uppercase_required
    ) {
      password += pickRandChar(source, CharListAlphabet);
    }
    if (policy.digit_required) {
      password += pickRandChar(source, CharListDigit);
    }
    if (policy.symbol_required) {
      password += pickRandChar(source, CharListSymbol);
    }

    for (let i = password.length; i < minLength; i++) {
      password += pickRandChar(source, charList);
    }

    password = source.shuffle(password);

    if (validatePassword(password, policy) === null) {
      return password;
    }
  }
  return null;
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
  let charList = "";
  set.forEach((cs) => {
    charList += cs;
  });

  return charList;
}

export function determineMinLength(
  passwordPolicy: PasswordPolicyConfig
): number {
  let minLength = passwordPolicy.min_length ?? DefaultMinLength;

  if (minLength < DefaultMinLength) {
    minLength = DefaultMinLength;
  }

  if (
    (passwordPolicy.minimum_guessable_level ?? 0) > 0 &&
    minLength < GuessableEnabledMinLength
  ) {
    minLength = GuessableEnabledMinLength;
  }

  return minLength;
}

function pickRandChar(source: RandSource, charList: string): string {
  const randomBytes = source.randomBytes(1);
  const maxByte = 256;
  const discard = maxByte - (maxByte % charList.length);
  let byte = randomBytes[0];

  while (byte >= discard) {
    byte = source.randomBytes(1)[0];
  }

  return charList[byte % charList.length];
}
