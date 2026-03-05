/**
 * @jest-environment node
 */
import { RuleTester } from "eslint";
import * as parser from "@typescript-eslint/parser";
import rule from "./no-unsafe-react-event-usage.cjs";

const ruleTester = new RuleTester({
  languageOptions: {
    parser,
    parserOptions: {
      ecmaVersion: 2022,
      sourceType: "module",
      ecmaFeatures: {
        jsx: true,
      },
    },
  },
});

ruleTester.run("no-unsafe-react-event-usage", rule, {
  valid: [
    {
      code: `
const onChangeSearchKeyword = (e: React.ChangeEvent<HTMLInputElement>) => {
  const value = e.currentTarget.value;
  onFilterChange((prev) => ({
    ...prev,
    searchKeyword: value,
  }));
};
`,
      filename: "test.tsx",
    },
    {
      code: `
const onClick = (event: React.MouseEvent<unknown>) => {
  event.preventDefault();
  event.stopPropagation();
  setState((prev) => ({ ...prev, enabled: true }));
};
`,
      filename: "test.tsx",
    },
    {
      code: `
const onChange = (_: unknown, newValue?: string) => {
  setState((prev) => ({ ...prev, value: newValue ?? "" }));
};
`,
      filename: "test.tsx",
    },
  ],
  invalid: [
    {
      code: `
const onChangeSearchKeyword = (e: React.ChangeEvent<HTMLInputElement>) => {
  onFilterChange((prev) => ({
    ...prev,
    searchKeyword: e.currentTarget.value,
  }));
};
`,
      filename: "test.tsx",
      errors: [{ messageId: "unsafeEventReference" }],
    },
    {
      code: `
const onClick = (event: React.MouseEvent<unknown>) => {
  Promise.resolve().then(() => {
    event.preventDefault();
  });
};
`,
      filename: "test.tsx",
      errors: [{ messageId: "unsafeEventReference" }],
    },
  ],
});
