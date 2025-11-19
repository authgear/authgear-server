import React from "react";
import type { Preview } from "@storybook/react-vite";
import { ThemeProvider } from "../src/components/v2/ThemeProvider/ThemeProvider";
import { AppLocaleProvider } from "../src/components/common/AppLocaleProvider";
import "../src/index.css";
import "@fortawesome/fontawesome-free/css/all.min.css";

const preview: Preview = {
  parameters: {
    layout: "centered",
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
  decorators: [
    (Story) => {
      return (
        <AppLocaleProvider>
          <ThemeProvider>
            <Story />
          </ThemeProvider>
        </AppLocaleProvider>
      );
    },
  ],
};

export default preview;
