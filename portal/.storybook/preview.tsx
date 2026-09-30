import React, { useEffect } from "react";
import type { Preview } from "@storybook/react-vite";
import { ThemeProvider } from "../src/components/v2/ThemeProvider/ThemeProvider";
import { AppLocaleProvider } from "../src/components/common/AppLocaleProvider";
import { parseAppearance, setAppearance } from "../src/util/appearance";
import "../src/index.css";
import "@fortawesome/fontawesome-free/css/all.min.css";
/** Tabler icon font for `<i className="ti ti-…">` (e.g. login method configuration). */
import "@tabler/icons/iconfont/tabler-icons.min.css";

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
  globalTypes: {
    appearance: {
      description: "Portal appearance",
      toolbar: {
        title: "Appearance",
        icon: "mirror",
        items: [
          { value: "light", title: "Light", icon: "sun" },
          { value: "dark", title: "Dark", icon: "moon" },
          { value: "system", title: "Device", icon: "browser" },
        ],
        dynamicTitle: true,
      },
    },
  },
  initialGlobals: {
    appearance: "light",
  },
  decorators: [
    (Story, context) => {
      const appearance = parseAppearance(context.globals.appearance);
      // Same code path as the portal: puts the light/dark class on <html>
      // and updates every useAppearance() consumer in the story.
      useEffect(() => {
        setAppearance(appearance);
      }, [appearance]);
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
