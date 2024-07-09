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
      height: (theme) => theme("spacing"),
      width: (theme) => theme("spacing"),
      minHeight: (theme) => theme("spacing"),
      maxHeight: (theme) => theme("spacing"),
      minWidth: (theme) => theme("spacing"),
      maxWidth: (theme) => theme("spacing"),
      spacing: () => {
        const spacing = {};
        for (let i = 0; i <= 300; i += 0.5) {
          spacing[i] = `${i * 0.25}rem`;
        }
        return spacing;
      },
    },
    screens: {
      tablet: "640px",
      desktop: "1024px",
    },
  },
  plugins: [],
};
