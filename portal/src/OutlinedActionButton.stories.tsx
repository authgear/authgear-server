import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import OutlinedActionButton from "./components/common/OutlinedActionButton";
import {
  defaultSystemConfig,
  instantiateSystemConfig,
} from "./system-config";

const { themes } = instantiateSystemConfig(defaultSystemConfig);

/**
 * Outlined Fluent default button; `theme` sets border/label `themePrimary` (e.g. `themes.destructive` on Account status).
 */
const meta = {
  title: "components/v1/Button/OutlinedActionButton",
  component: OutlinedActionButton,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          boxSizing: "border-box",
          minHeight: "70vh",
          width: "100%",
          padding: 24,
        }}
      >
        <Story />
      </div>
    ),
  ],
  args: {
    theme: themes.destructive,
    text: "Save changes",
    onClick: () => {},
  },
} satisfies Meta<typeof OutlinedActionButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};

export const WithIcon: Story = {
  args: {
    iconProps: { iconName: "Save" },
  },
};
