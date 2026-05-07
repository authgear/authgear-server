import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import ActionButton from "./ActionButton";

const meta = {
  title: "components/v1/Button/ActionButton",
  component: ActionButton,
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
    text: "Add item",
    iconProps: { iconName: "Add" },
  },
} satisfies Meta<typeof ActionButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};
