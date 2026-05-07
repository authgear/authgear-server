import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { MessageBar, MessageBarType, ThemeProvider } from "@fluentui/react";
import RedMessageBar from "./RedMessageBar";
import BlueMessageBar from "./BlueMessageBar";

export type PortalMessageBarVariant = "Primary" | "Error" | "Warning";

export interface PortalMessageBarStoryProps {
  /** Primary → `BlueMessageBar`; Error → `RedMessageBar`; Warning → Fluent `MessageBar`. */
  variant: PortalMessageBarVariant;
  isMultiline?: boolean;
  children: React.ReactNode;
}

function renderPortalMessageBar(
  props: PortalMessageBarStoryProps
): React.ReactElement {
  const { variant, children } = props;
  /** Storybook controls must be a real boolean — Fluent treats any truthy value as multiline. */
  const isMultiline = props.isMultiline !== false;

  switch (variant) {
    case "Primary":
      return (
        <BlueMessageBar isMultiline={isMultiline}>{children}</BlueMessageBar>
      );
    case "Error":
      return (
        <RedMessageBar isMultiline={isMultiline}>{children}</RedMessageBar>
      );
    case "Warning":
      return (
        <MessageBar
          messageBarType={MessageBarType.warning}
          isMultiline={isMultiline}
        >
          {children}
        </MessageBar>
      );
  }
}

const meta = {
  title: "components/v1/MessageBar",
  tags: ["autodocs"],
  /** No `component` — avoids Storybook adding an extra sidebar entry from the render function name. */
  render: (args: PortalMessageBarStoryProps) => renderPortalMessageBar(args),
  decorators: [
    (Story) => (
      <ThemeProvider>
        <div style={{ width: 480 }}>
          <Story />
        </div>
      </ThemeProvider>
    ),
  ],
  argTypes: {
    variant: {
      control: "select",
      options: [
        "Primary",
        "Error",
        "Warning",
      ] satisfies PortalMessageBarVariant[],
      description:
        "Primary → BlueMessageBar, Error → RedMessageBar, Warning → Fluent MessageBar",
    },
    children: {
      control: "text",
    },
    isMultiline: {
      control: "boolean",
    },
  },
  args: {
    variant: "Primary",
    isMultiline: true,
    children:
      "Informational notice using primary-tint styling (e.g. quota or feature hints). This sentence is long enough to wrap inside a 480px-wide message bar when multiline is on.",
  },
} satisfies Meta<PortalMessageBarStoryProps>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: {
    variant: "Primary",
    children:
      "Informational notice using primary-tint styling (e.g. quota or feature hints). This sentence is long enough to wrap inside a 480px-wide message bar when multiline is on.",
  },
};

export const Error: Story = {
  args: {
    variant: "Error",
    children:
      "Something went wrong. Check your configuration and try again. Add more detail here so the line wraps when multiline is enabled — Fluent uses single-line ellipsis when multiline is off.",
  },
};

export const Warning: Story = {
  args: {
    variant: "Warning",
    children:
      "You have unsaved changes. Save or discard before leaving this screen. With multiline on and a narrow width, this copy should wrap to several lines instead of staying on one row.",
  },
};
