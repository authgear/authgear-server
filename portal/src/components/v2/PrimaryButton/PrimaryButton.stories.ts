import type { Meta, StoryObj } from "@storybook/react";
import { PrimaryButton } from "./PrimaryButton";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: PrimaryButton,
  tags: ["autodocs"],
  argTypes: {
    text: {
      control: {
        type: "text",
      },
    },
  },
  args: {
    text: "Start",
  },
} satisfies Meta<typeof PrimaryButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    size: "2",
    highContrast: false,
    disabled: false,
  },
};

export const Dark: Story = {
  args: {
    darkMode: true,
    size: "2",
    highContrast: true,
    disabled: false,
  },
  parameters: {
    backgrounds: {
      default: "dark",
    },
  },
};
