module.exports = {
  syntax: "postcss-scss",
  plugins: [require("@csstools/postcss-sass")({}), require("autoprefixer")({})],
};
