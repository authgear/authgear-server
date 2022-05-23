module.exports = {
  purge: ["../resources/authgear/templates/en/web/**/*.html"],
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
  variants: {
    extend: {},
  },
  plugins: [],
};
