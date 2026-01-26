import oursky from "@oursky/eslint-plugin";
import globals from "globals";

const js = "src/**/*.{js,jsx,mjs,mjsx}";
const ts = "src/**/*.{ts,tsx,mts,mtsx}";

export default [
  {
    ignores: ["src/**/*.generated.*", "**/dist/"],
  },
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
    files: [ts],
    ...oursky.configs.typescript,
  },
  {
    files: [ts],
    ...oursky.configs.tsdoc,
  },
  {
    files: [js, ts],
    ...oursky.configs.react,
  },
  {
    files: [js, ts],
    ...oursky.configs["react-hooks"],
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
      complexity: "off",
      "sonarjs/cognitive-complexity": "off",
      "tsdoc/syntax": "off",
      "react/jsx-no-bind": "off",
      "react/forbid-elements": [
        "error",
        { forbid: ["h1", "h2", "h3", "h4", "h5", "h6"] },
      ],
      "react/forbid-component-props": [
        "error",
        {
          forbid: [
            {
              propName: "subText",
              allowedFor: [],
              message: "subText is deprecated in Dialog component",
            },
          ],
        },
      ],
      // We have some places to Boolean(b1 operator b2) to work around react/jsx-* rule
      // So allow it.
      "@typescript-eslint/no-unnecessary-type-conversion": "off",
      "@typescript-eslint/no-unsafe-enum-comparison": "off",
      // If this is turned on, we have over 300 errors. So it is turned off.
      "@typescript-eslint/no-unsafe-type-assertion": "off",
      // We have many places using default for the exhaustive checking. So allow it.
      "@typescript-eslint/switch-exhaustiveness-check": [
        "error",
        {
          considerDefaultExhaustiveForUnions: true,
        },
      ],
      "@typescript-eslint/use-unknown-in-catch-callback-variable": "off",
      "@typescript-eslint/no-floating-promises": "off",
      "@typescript-eslint/no-misused-promises": [
        "error",
        {
          checksVoidReturn: {
            attributes: false,
          },
        },
      ],
      "no-void": "off",
      "@typescript-eslint/no-non-null-assertion": "off",
      "@typescript-eslint/strict-boolean-expressions": "off",
      "@typescript-eslint/prefer-nullish-coalescing": "off",
      "@typescript-eslint/no-deprecated": "off",
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "@elgorditosalsero/react-gtm-hook",
              message:
                'Please import "GTMProvider" from ./src/GTMProvider instead.',
            },
            {
              name: "@fluentui/react",
              importNames: [
                "TextField",
                "PrimaryButton",
                "DefaultButton",
                "MessageBarButton",
                "ActionButton",
                "CommandBarButton",
                "Link",
                "Toggle",
              ],
              message: "Please import the replacement from ./src instead.",
            },
            {
              name: "zxcvbn",
              message: "Please import from ./src/util/zxcvbn instead.",
            },
            {
              name: "@fluentui/react",
              importNames: ["Pivot"],
              message:
                "Please use AGPivot from src/components/common/AGPivot instead.",
            },
          ],
        },
      ],
    },
  },
];
