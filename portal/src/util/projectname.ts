// ref https://github.com/bitcoinjs/bip39/blob/master/src/wordlists/english.json
import * as wordlist from "./wordlist.json";

const RANDOM_ALPHA_NUMERIC_STRING_LENGTH = 6;

function determineWord(index: number): string {
  return wordlist[index];
}

export function getRandom32BitsNumber(): number {
  const randomBuffer = new Uint32Array(1);
  window.crypto.getRandomValues(randomBuffer);
  return randomBuffer[0];
}

export function maskNumber(num: number, startAt: number, bits: number): number {
  return (num >> startAt) & ((1 << bits) - 1);
}

export function getRandomAlphaNumericString(len: number): string {
  let result = "";
  const chars = "abcdefghijklmnopqrstuvwxyz0123456789";
  for (let i = 0; i < len; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

export function deterministicProjectName(num: number): string {
  const firstRandomStringIndex = maskNumber(num, 21, 11);

  const firstRandomString = determineWord(firstRandomStringIndex);

  const randomAlphaNumericString = getRandomAlphaNumericString(
    RANDOM_ALPHA_NUMERIC_STRING_LENGTH
  );

  return `${firstRandomString}-${randomAlphaNumericString}`;
}

export function randomProjectName(): string {
  return deterministicProjectName(getRandom32BitsNumber());
}
