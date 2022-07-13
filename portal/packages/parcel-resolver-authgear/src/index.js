const { Resolver } = require("@parcel/plugin");
const { readFile } = require("fs/promises");

let packageJSON = null;

function specifierToPackage(specifier) {
  if (specifier.startsWith("@")) {
    const parts = specifier.split("/");
    return parts.slice(0, 2).join("/");
  }
  const parts = specifier.split("/");
  return parts[0];
}

const POLYFILL_PREFIX = ["@swc/helpers", "@parcel/transformer-"];

function isPolyfill(package) {
  for (const prefix of POLYFILL_PREFIX) {
    if (package.startsWith(prefix)) {
      return true;
    }
  }
  return false;
}

// The purpose of this plugin is to prevent our JavaScript files
// from importing unlisted dependencies.
// If such thing happen, code splitted bundle will NOT import
// the unlisted dependencies, causing module not found error at runtime.
module.exports = new Resolver({
  async resolve(info) {
    if (packageJSON == null) {
      packageJSON = JSON.parse(
        await readFile("./package.json", { encoding: "utf8" })
      );
    }

    const {
      specifier,
      dependency: { sourcePath, sourceAssetType },
    } = info;
    // It is our source file
    if (
      sourcePath != null &&
      !sourcePath.includes("node_modules") &&
      sourceAssetType === "js"
    ) {
      // importing packages from node_modules
      if (specifier != null && !specifier.startsWith(".")) {
        const package = specifierToPackage(specifier);
        const polyfill = isPolyfill(package);
        // and it is not polyfill.
        if (!polyfill) {
          // Make sure the dependency is listed in package.json
          const version = packageJSON["dependencies"][package];
          if (version == null) {
            throw new Error(
              `${sourcePath} imports ${specifier} which is unlisted in package.json`
            );
          }
        }
      }
    }
    return null;
  },
});
