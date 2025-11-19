import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { StepTrail } from "./Stepper";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: StepTrail,
  tags: ["autodocs"],
  args: {
    progress: 0.5,
  },
  render: (args) => {
    return (
      <div style={{ width: "80px" }}>
        <StepTrail {...args} />
      </div>
    );
  },
} satisfies Meta<typeof StepTrail>;

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
