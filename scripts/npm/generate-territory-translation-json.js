const { argv } = require("process");
// Valid values are en, zh-Hant-HK, zh-Hant.
const lang = argv[2];
const importPath = `cldr-localenames-modern/main/${lang}/territories.json`;
const data = require(importPath);

const alpha2ToLocalizedName = data.main[lang].localeDisplayNames.territories;

const output = {};
for (const [maybeAlpha2, value] of Object.entries(alpha2ToLocalizedName)) {
  if (/^[A-Z]{2}$/.test(maybeAlpha2)) {
    output[`territory-${maybeAlpha2}`] = value;
  }
}
console.log(JSON.stringify(output, null, 2));
