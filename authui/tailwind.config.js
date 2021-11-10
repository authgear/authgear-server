module.exports = {
  purge: ["../resources/authgear/templates/en/web/**/*.html"],
  darkMode: false, // or 'media' or 'class'
  theme: {
    extend: {},
    screens: {
      tablet: "640px",
      desktop: "1024px",
    },
  },
  variants: {
    extend: {},
  },
  plugins: [],
};
