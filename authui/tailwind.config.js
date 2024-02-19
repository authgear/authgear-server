/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../resources/authgear/templates/en/web/**/*.html",
    "./src/**/*.{css,scss,ts}",
  ],
  // We do not actually mention the oauth provider types in the HTML.
  // So we have to safelist them here to ensure we generate the CSS.
  safelist: [
    "apple",
    "google",
    "facebook",
    "github",
    "linkedin",
    "azureadv2",
    "azureadb2c",
    "adfs",
    "wechat",
  ],
  darkMode: "class",
  theme: {
    extend: {
      flex: {
        "1-0-auto": "1 0 auto",
      },
      spacing: {
        18: "4.5rem",
      },
      maxWidth: {
        90: "90rem",
      },
    },
    screens: {
      tablet: "640px",
      desktop: "1024px",
    },
  },
  plugins: [],
};
