import data from "tzdata";

export const TIMEZONE_NAMES = (() => {
  const names = [];
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

    names.push(key);
  }

  names.sort((a, b) => {
    return a.localeCompare(b);
  });

  return names;
})();
