import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import DefaultButton from "./DefaultButton";

const meta = {
  title: "components/v1/Button/DefaultButton",
  component: DefaultButton,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          boxSizing: "border-box",
          width: "100%",
          padding: 16,
        }}
      >
        <Story />
      </div>
    ),
  ],
  args: {
    text: "Cancel",
  },
} satisfies Meta<typeof DefaultButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ThemePrimaryBorder: Story = {
  args: {
    useThemePrimaryForBorderColor: true,
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};
