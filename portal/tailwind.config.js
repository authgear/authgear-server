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
      flex: {
        1: "1 1 0", // The default 1 1 0% doesn't work in some cases, change to 0
        "1-0-auto": "1 0 auto",
      },
      colors: {
        grey: { white7: "#F4F4F4" },
        status: {
          green: "#10B070",
          grey: "#605E5C",
        },
        theme: {
          primary: "#176df3",
        },
        neutral: {
          light: "#edebe9",
          lighter: "#f3f2f1",
        },
        separator: "#EDEBE9",
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
  },
  plugins: [require("@savvywombat/tailwindcss-grid-areas")],
};
