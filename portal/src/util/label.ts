// generateLabel assume val is in English, and generate
// a label suitable for layman.
export function generateLabel(val: string): string {
  const parts = val.split("_");
  const words: string[] = parts.map((word, index) => {
    return titlecase(word, index, parts.length);
  });
  return words.join(" ");
}

const MINOR_WORDS = [
  "and",
  "but",
  "for",
  "or",
  "nor",
  "the",
  "a",
  "an",
  "to",
  "as",
];

function titlecase(word: string, index: number, length: number): string {
  // The rules used here is a simplified version of the following style.
  // https://en.wikipedia.org/wiki/Title_case#Chicago_Manual_of_Style
  const lowercase = word.toLowerCase();

  let shouldCapitalize: boolean;
  if (index === 0 || index === length - 1) {
    shouldCapitalize = true;
  } else if (MINOR_WORDS.indexOf(lowercase) >= 0) {
    shouldCapitalize = false;
  } else {
    shouldCapitalize = true;
  }

  if (shouldCapitalize) {
    const chars = [...lowercase];
    chars[0] = chars[0].toUpperCase();
    return chars.join("");
  }

  return lowercase;
}
