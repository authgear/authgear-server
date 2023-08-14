module.exports = {
  corePlugins: {
    preflight: false,
  },
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  darkMode: "class",
  theme: {
    screens: {
      mobile: { max: "640px" },
      tablet: { max: "1080px" },
    },
    extend: {
      colors: {
        grey: { white7: "#F4F4F4" },
        status: {
          green: "#33BA89",
          grey: "#595653",
        },
      },
    },
  },
  plugins: [require("@savvywombat/tailwindcss-grid-areas")],
};
