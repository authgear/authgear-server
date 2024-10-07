import oursky from "@oursky/eslint-plugin";
import compat from "eslint-plugin-compat";
import globals from "globals";

// FIXME(eslint): For now we only lint src/authflowv2.
// but we should lint everything in src.
const js = "src/authflowv2/**/*.{js,jsx,mjs,mjsx}";
const ts = "src/authflowv2/**/*.{ts,tsx,mts,mtsx}";

export default [
  {
    files: [ts],
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    files: [js, ts],
    ...oursky.configs.eslint,
  },
  {
    files: [js, ts],
    ...compat.configs["flat/recommended"],
  },
  {
    files: [ts],
    ...oursky.configs.typescript,
  },
  {
    files: [ts],
    ...oursky.configs.tsdoc,
  },
  {
    files: [js, ts],
    ...oursky.configs.oursky,
  },
  {
    files: [js, ts],
    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },
  },
  {
    files: [js, ts],
    rules: {
      "no-console": ["error", { allow: ["warn", "error"] }],
      "no-void": "off",
      "@typescript-eslint/class-methods-use-this": "off",
      "@typescript-eslint/explicit-module-boundary-types": "off",
      "@typescript-eslint/strict-boolean-expressions": "off",
      "tsdoc/syntax": "off",
    },
  },
];
