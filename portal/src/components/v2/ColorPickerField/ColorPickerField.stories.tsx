import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { ColorPickerField, ColorPickerFieldProps } from "./ColorPickerField";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: ColorPickerField,
  tags: ["autodocs"],
  argTypes: {
    error: {
      control: {
        type: "text",
      },
    },
    hint: {
      control: {
        type: "text",
      },
    },
  },
  args: {
    size: "3",
    label: "Label",
    value: "",
  },
  render: (args) => {
    return <Demo {...args} />;
  },
} satisfies Meta<typeof ColorPickerField>;

function Demo({ value: _, ...args }: ColorPickerFieldProps) {
  const [value, setValue] = useState("#176DF3");

  return (
    <div style={{ width: "300px" }}>
      <ColorPickerField value={value} onValueChange={setValue} {...args} />
    </div>
  );
}

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
