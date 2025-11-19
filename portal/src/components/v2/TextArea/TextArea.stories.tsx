import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { TextArea } from "./TextArea";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: TextArea,
  tags: ["autodocs"],
  argTypes: {
    error: {
      control: {
        type: "text",
      },
    },
  },
  args: {
    size: "3",
    label: "Label",
    placeholder: "Placeholder",
  },
  render: (args) => {
    return (
      <div style={{ width: "300px" }}>
        <TextArea {...args} />
      </div>
    );
  },
} satisfies Meta<typeof TextArea>;

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

export const Error: Story = {
  args: {
    error: "This field is required",
  },
};
