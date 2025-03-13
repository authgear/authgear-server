import React from "react";
import type { Preview } from "@storybook/react";
import { Theme } from "@radix-ui/themes";
import "../src/index.css";

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
  decorators: [
    (Story) => (
      <Theme className="contents">
        <Story />
      </Theme>
    ),
  ],
};

export default preview;
