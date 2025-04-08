import React from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { ArrowLeftIcon } from "@radix-ui/react-icons";
import { TextButton } from "./TextButton";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: TextButton,
  tags: ["autodocs"],
  argTypes: {
    text: {
      control: {
        type: "text",
      },
    },
  },
  args: {
    text: "Click Me",
    size: "2",
  },
} satisfies Meta<typeof TextButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    variant: "default",
    darkMode: false,
    disabled: false,
  },
};

export const DarkSecondary: Story = {
  args: {
    variant: "secondary",
    darkMode: true,
    disabled: false,
  },
  parameters: {
    backgrounds: {
      default: "dark",
    },
  },
};

export const DarkSecondaryBack: Story = {
  args: {
    variant: "secondary",
    darkMode: true,
    disabled: false,
    iconStart: <ArrowLeftIcon width={20} height={20} />,
    text: "Back",
  },
  parameters: {
    backgrounds: {
      default: "dark",
    },
  },
};
