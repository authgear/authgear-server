import React from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { TextField } from "./TextField";
import { InfoCircledIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: TextField,
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
        <TextField {...args} />
      </div>
    );
  },
} satisfies Meta<typeof TextField>;

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

export const IconsStartEnd: Story = {
  args: {
    iconStart: <MagnifyingGlassIcon height="16" width="16" />,
    iconEnd: <InfoCircledIcon height="16" width="16" />,
  },
};
