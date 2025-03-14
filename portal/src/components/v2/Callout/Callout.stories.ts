import type { Meta, StoryObj } from "@storybook/react";
import { Callout } from "./Callout";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  title: "common/Callout",
  component: Callout,
  parameters: {
    layout: "centered",
  },
  tags: ["autodocs"],
  argTypes: {
    text: {
      control: {
        type: "text",
      },
    },
    showCloseButton: {
      control: {
        type: "boolean",
      },
    },
  },
  args: {
    text: "Some text in the callout",
    showCloseButton: true,
  },
} satisfies Meta<typeof Callout>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Error: Story = {
  args: {
    type: "error",
  },
};

export const Success: Story = {
  args: {
    type: "success",
  },
};
