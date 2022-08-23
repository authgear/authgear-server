module.exports = {
  corePlugins: {
    preflight: false,
  },
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  darkMode: "class",
  theme: {
    screens: {
      mobile: { max: "640px" },
      desktop: "640px",
    },
  },
  plugins: [require("@savvywombat/tailwindcss-grid-areas")],
};
