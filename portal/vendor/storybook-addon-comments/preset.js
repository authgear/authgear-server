import { createRequire } from "node:module";

const require = createRequire(import.meta.url);

/** @param {string[]} [entry] */
export function managerEntries(entry = []) {
  return [...entry, require.resolve("./src/manager.tsx")];
}
