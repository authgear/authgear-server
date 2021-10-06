import { getAlpha2Codes, getName } from "i18n-iso-countries";

export interface Option {
  key: string;
  text: string;
}

export function makeAlpha2Options(): Option[] {
  const map = getAlpha2Codes();
  const alpha2 = Object.keys(map);
  const options = alpha2.map((a) => {
    return {
      key: a,
      text: getName(a, "en"),
    };
  });
  options.sort((o1, o2) => {
    return o1.text.localeCompare(o2.text);
  });
  return options;
}
