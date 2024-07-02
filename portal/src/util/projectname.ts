// ref https://github.com/bitcoinjs/bip39/blob/master/src/wordlists/english.json
import * as wordlist from "./wordlist.json";

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

export function deterministicProjectName(num: number): string {
  const randomNumber = maskNumber(num, 0, 10);
  const secondRandomStringIndex = maskNumber(num, 10, 11);
  const firstRandomStringIndex = maskNumber(num, 21, 11);

  const firstRandomString = determineWord(firstRandomStringIndex);
  const secondRandomString = determineWord(secondRandomStringIndex);

  return `${firstRandomString}-${secondRandomString}-${randomNumber}`;
}

export function randomProjectName(): string {
  return deterministicProjectName(getRandom32BitsNumber());
}
