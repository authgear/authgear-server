import type { Meta, StoryObj } from "@storybook/react-vite";
import { Callout } from "./Callout";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: Callout,
  tags: ["autodocs"],
  argTypes: {
    text: {
      control: {
        type: "text",
      },
    },
  },
  args: {
    className: "w-90",
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

export const Warning: Story = {
  args: {
    type: "warning",
  },
};

export const Success: Story = {
  args: {
    type: "success",
  },
};
