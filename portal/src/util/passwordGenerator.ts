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

const AllCharLists = [
  CharListLowercase,
  CharListUppercase,
  CharListAlphabet,
  CharListDigit,
  CharListSymbol,
];

export const MaxTrials = 10;
const DefaultMinLength = 8;
const GuessableEnabledMinLength = 32;

export interface RandSource {
  intN(n: number): number;
}

export const cryptoRandSource: RandSource = {
  intN: (n: number): number => {
    const MAX_SAFE_INTEGER = Math.pow(2, 32) - 1; // uint32
    if (n > MAX_SAFE_INTEGER) {
      throw new Error("n must be less than 2^32");
    }

    const randomArray = new Uint8Array(Math.max(4, n));
    const dataView = new DataView(randomArray.buffer);

    // Generate a random number in [0, n).
    // It rejects numbers that are greater than or equal to
    // Math.floor(MAX_SAFE_INTEGER / n) * n to avoid modulo bias.
    let randomNumber;
    do {
      window.crypto.getRandomValues(randomArray);
      randomNumber = dataView.getUint32(0, true);
    } while (randomNumber >= Math.floor(MAX_SAFE_INTEGER / n) * n);

    return randomNumber % n;
  },
};

export const mathRandSource: RandSource = {
  intN: (n: number): number => {
    return Math.floor(Math.random() * n);
  },
};

export function generatePassword(policy: PasswordPolicyConfig): string | null {
  try {
    window.crypto.getRandomValues(new Uint8Array(1));
    return generatePasswordWithSource(cryptoRandSource, policy, MaxTrials);
  } catch {
    return generatePasswordWithSource(mathRandSource, policy, MaxTrials);
  }
}

export function generatePasswordWithSource(
  source: RandSource,
  policy: PasswordPolicyConfig,
  maxTrials: number
): string | null {
  const [password, _] = internalGeneratePasswordWithSource(
    source,
    policy,
    maxTrials
  );
  return password;
}

export function internalGeneratePasswordWithSource(
  source: RandSource,
  policy: PasswordPolicyConfig,
  maxTrials: number
): [string | null, number] {
  for (let i = 0; i < maxTrials; i++) {
    const password = generatePasswordOnce(source, policy);
    if (password !== null) {
      return [password, i];
    }
  }

  return [null, -1];
}

// eslint-disable-next-line complexity
function generatePasswordOnce(
  source: RandSource,
  policy: PasswordPolicyConfig
): string | null {
  const minLength = determineMinLength(policy);
  const charList = prepareCharList(policy);

  const passwordArray: string[] = [];

  // Add required characters.
  if (policy.lowercase_required) {
    passwordArray.push(pickRandChar(source, CharListLowercase));
  }
  if (policy.uppercase_required) {
    passwordArray.push(pickRandChar(source, CharListUppercase));
  }
  if (
    policy.alphabet_required &&
    !policy.lowercase_required &&
    !policy.uppercase_required
  ) {
    passwordArray.push(pickRandChar(source, CharListAlphabet));
  }
  if (policy.digit_required) {
    passwordArray.push(pickRandChar(source, CharListDigit));
  }
  if (policy.symbol_required) {
    passwordArray.push(pickRandChar(source, CharListSymbol));
  }

  // Fill the rest of the password with random characters.
  for (let i = passwordArray.length; i < minLength; i++) {
    passwordArray.push(pickRandChar(source, charList));
  }

  // Shuffle the password since we have required characers at the beginning.
  for (let i = passwordArray.length - 1; i > 0; i--) {
    const j = source.intN(i + 1);
    [passwordArray[i], passwordArray[j]] = [passwordArray[j], passwordArray[i]];
  }

  const password = passwordArray.join("");
  if (validatePassword(password, policy) === null) {
    return password;
  }

  return null;
}

// eslint-disable-next-line complexity
export function prepareCharList(passwordPolicy: PasswordPolicyConfig): string {
  const set = new Set<string>();

  // Use alphanumeric as base character set
  set.add(CharListAlphabet);
  set.add(CharListDigit);

  if (passwordPolicy.symbol_required) set.add(CharListSymbol);

  // Build the final character list.
  let charList = "";
  for (const cs of AllCharLists) {
    if (set.has(cs)) {
      charList += cs;
    }
  }

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
  return charList[source.intN(charList.length)];
}
