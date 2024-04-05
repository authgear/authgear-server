export function cleanRawInputValue(rawInputValue: string): string {
  const digits = ["0", "1", "2", "3", "4", "5", "6", "7", "8", "9"];
  let out = "";
  let i = 0;
  for (const rune of rawInputValue) {
    if (i === 0) {
      if (rune === "+") {
        out += rune;
      }
    }

    if (digits.includes(rune)) {
      out += rune;
    }

    i += 1;
  }
  return out;
}

export function trimCountryCallingCode(
  rawInputValue: string,
  countryCallingCode: string
): string {
  if (rawInputValue === "+") {
    return "";
  }
  const prefix = `+${countryCallingCode}`;
  if (rawInputValue.startsWith(prefix)) {
    return rawInputValue.slice(prefix.length);
  }
  return rawInputValue;
}

export function makePartialValue(
  rawInputValue: string,
  countryCallingCode: string
): string {
  const trimmed = trimCountryCallingCode(rawInputValue, countryCallingCode);
  return `+${countryCallingCode}${trimmed}`;
}
