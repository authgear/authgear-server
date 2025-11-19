import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { Stepper } from "./Stepper";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: Stepper,
  tags: ["autodocs"],
  args: {
    steps: [
      {
        text: "1",
        checked: true,
      },
      {
        text: "2",
        checked: true,
      },
      {
        text: "3",
        checked: true,
      },
      {
        text: "4",
        checked: false,
      },
      {
        text: "5",
        checked: false,
      },
    ],
  },
  render: (args) => {
    return (
      <div style={{ width: "460px" }}>
        <Stepper {...args} />
      </div>
    );
  },
} satisfies Meta<typeof Stepper>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {},
};

export const Dark: Story = {
  args: {
    darkMode: true,
  },
  parameters: {
    backgrounds: {
      default: "dark",
    },
  },
};
