module.exports = {
  content: ["../resources/authgear/templates/en/web/**/*.html"],
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
