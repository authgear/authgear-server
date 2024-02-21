// Ref: https://github.com/annexare/Countries/blob/d92daf336b0d71cf265c5689e7a2b789832bf3bc/packages/countries/src/getEmojiFlag.ts

// "Regional Indicator Symbol Letter A" - "Latin Capital Letter A"
const UNICODE_BASE = 127462 - "A".charCodeAt(0);

// Country code should contain exactly 2 uppercase characters from A..Z
const COUNTRY_CODE_REGEX = /^[A-Z]{2}$/;

export function getEmojiFlag(countryCode: string): string {
  if (COUNTRY_CODE_REGEX.test(countryCode)) {
    return String.fromCodePoint(
      ...countryCode
        .split("")
        .map((letter) => UNICODE_BASE + letter.toUpperCase().charCodeAt(0))
    );
  }

  return "";
}
