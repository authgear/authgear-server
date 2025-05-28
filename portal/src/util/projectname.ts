// ./wordlist.json is originally https://github.com/bitcoinjs/bip39/blob/master/src/wordlists/english.json
// However, we found that some of the words are actually reserved words in reserved_app_id.txt
// Therefore, the following words are removed from the list.
// The implementation is also changed to accommodate the fact that the list is no longer 2048 items long.
// Removed words: about, abuse, access, account, alpha, apple, arrow, asset, atom, auto, ball, board, book, business, chat, client, cloud, code, connect, control, design, destroy, develop, direct, dish, domain, edit, example, exchange, favorite, feed, file, forum, friend, gallery, game, gate, general, gift, guide, help, home, host, icon, image, index, invite, job, lab, legal, library, link, list, live, local, mail, main, manage, market, master, media, member, mirror, mobile, monitor, net, news, office, old, online, panel, payment, photo, picture, post, private, project, public, put, report, search, service, share, shop, social, sport, staff, stage, start, student, style, survey, system, tag, target, team, test, ticket, time, tool, track, travel, update, upgrade, video, visa, web, work
import wordlist from "./wordlist.json";

function determineWord(seed: number): string {
  const index = seed % wordlist.length;
  return wordlist[index];
}

export function getRandom48Bits(): Uint8Array {
  const randomBuffer = new Uint8Array(6);
  window.crypto.getRandomValues(randomBuffer);
  return randomBuffer;
}

/**
 * This function take 48 bits, and split into 2 numbers, bit_0_10 and bit_11_41, dropping the last 6 bits
 *
 * @export
 * @param {Uint8Array} fortyEightBits
 * @returns {[number, number]}
 */
export function extractBits(fortyEightBits: Uint8Array): [number, number] {
  const bit_0_7 = fortyEightBits[0];
  const bit_8_15 = fortyEightBits[1];
  const bit_16_23 = fortyEightBits[2];
  const bit_24_31 = fortyEightBits[3];
  const bit_32_39 = fortyEightBits[4];
  const bit_40_47 = fortyEightBits[5];

  // 11 bits are needed, there are 2^11 = 2048 words in the wordlist
  const bit_0_10 = (bit_0_7 << 3) | (bit_8_15 >>> 5);

  const bit_11_15 = bit_8_15 & 0b00011111;
  const bit_40_41 = bit_40_47 >>> 6;

  // 31 bits are needed. We need 6 lowercase-alphanumeric characters.
  // There are 36 lowercase-alphanumeric characters in total.
  // log36(2^31) = 5.996...
  const bit_11_41 =
    (bit_11_15 << 26) |
    (bit_16_23 << 18) |
    (bit_24_31 << 10) |
    (bit_32_39 << 2) |
    bit_40_41;

  return [bit_0_10, bit_11_41];
}

export function deterministicAlphanumericString(bits: number): string {
  const string = bits.toString(36);

  if (string.length > 6) {
    throw new Error("number of bits must be less than 31");
  }

  const zeroPaddedString = string.padStart(6, "0");
  return zeroPaddedString;
}

/**
 * This function take 48 bits
 * The  1st - 11th bits bit_0_10 are used to determine the word
 * The 12th - 42nd bits bit_11_41 are used to determine the alphanumeric string
 * The 43rd - 48th bits bit_42_47 are not used
 *
 * @export
 * @param {Uint8Array} fortyEightBits Uint8Array of length 6
 * @returns {string}
 */
export function deterministicProjectID(fortyEightBits: Uint8Array): string {
  if (fortyEightBits.length !== 6) {
    throw new Error("fortyEightBits must be 6 bytes");
  }
  const [bit_0_10, bit_11_41] = extractBits(fortyEightBits);
  const firstRandomString = determineWord(bit_0_10);

  const alphaNumericString = deterministicAlphanumericString(bit_11_41);

  return `${firstRandomString}-${alphaNumericString}`;
}

export function randomProjectID(): string {
  return deterministicProjectID(getRandom48Bits());
}

export function projectIDFromCompanyName(companyName: string): string {
  const randomBits = getRandom48Bits();
  const [_, thirtyOneBits] = extractBits(randomBits);
  const alphaNumericString = deterministicAlphanumericString(thirtyOneBits);
  return `${companyName}-${alphaNumericString}`;
}

export function processCompanyName(companyName: string): string {
  return companyName
    .trim()
    .split("")
    .filter((char) => /[a-zA-Z\s]/.exec(char))
    .join("")
    .split(" ")
    .filter((word) => word !== "")
    .join("-")
    .toLowerCase();
}
