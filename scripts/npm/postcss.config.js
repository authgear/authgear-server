var presetEnv = require("postcss-preset-env");
var normalize = require("postcss-normalize");
var cssnano = require("cssnano");

module.exports = {
  plugins: [
    presetEnv({
      browsers: [
        "last 2 chrome versions",
        "last 2 firefox versions",
        "Firefox ESR",
        "IE >= 11",
        "iOS >= 11",
        "Android >= 5.0",
      ],
    }),
    normalize(),
    cssnano(),
  ],
};
