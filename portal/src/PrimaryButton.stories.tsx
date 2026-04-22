import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import PrimaryButton from "./PrimaryButton";

const meta = {
  title: "components/v1/Button/PrimaryButton",
  component: PrimaryButton,
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
    text: "Save changes",
  },
} satisfies Meta<typeof PrimaryButton>;

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
