import data from "tzdata";
import { IANAZone } from "luxon";

export interface Option {
  key: string;
  text: string;
  timezoneOffset: number;
}

export function makeTimezoneOptions(): Option[] {
  const options = [];
  const refTime = new Date().getTime();

  for (const [key, value] of Object.entries(data.zones)) {
    // This is an alias.
    if (typeof value === "string") {
      continue;
    }

    if (!key.includes("/")) {
      continue;
    }

    if (key.startsWith("Etc/")) {
      continue;
    }

    const iana = IANAZone.create(key);
    if (!iana.isValid) {
      continue;
    }

    const timezoneOffset = iana.offset(refTime);
    const text = `[UTC ${iana.formatOffset(refTime, "short")}] ${key}`;

    options.push({
      key,
      text,
      timezoneOffset,
    });
  }

  options.sort((a, b) => {
    return a.timezoneOffset - b.timezoneOffset;
  });

  return options;
}
