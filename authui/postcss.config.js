var presetEnv = require("postcss-preset-env");
var normalize = require("postcss-normalize");
var cssnano = require("cssnano");

module.exports = {
  plugins: [presetEnv(), normalize(), cssnano()]
};
