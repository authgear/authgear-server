// ref https://github.com/bitcoinjs/bip39/blob/master/src/wordlists/english.json
import * as wordlist from "./wordlist.json";

export function determineWord(index: number): string {
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

export function randomProjectName(): string {
  const random32BitsNumber = getRandom32BitsNumber();
  const firstRandomStringIndex = maskNumber(random32BitsNumber, 0, 11);
  const secondRandomStringIndex = maskNumber(random32BitsNumber, 11, 11);
  const randomNumber = maskNumber(random32BitsNumber, 22, 10);

  const firstRandomString = determineWord(firstRandomStringIndex);
  const secondRandomString = determineWord(secondRandomStringIndex);

  return `${firstRandomString}-${secondRandomString}-${randomNumber}`;
}
