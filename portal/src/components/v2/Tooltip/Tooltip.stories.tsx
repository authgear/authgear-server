import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { Tooltip } from "./Tooltip";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: Tooltip,
  tags: ["autodocs"],

  argTypes: {
    children: { table: { disable: true } },
  },
  args: {
    content: "This is a tooltip",
  },
  render: (args) => {
    return (
      <div style={{ width: "300px" }}>
        <Tooltip {...args}>
          <span>Hover Me</span>
        </Tooltip>
      </div>
    );
  },
} satisfies Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {},
};
