import type { Meta, StoryObj } from "@storybook/react-vite";
import { ImageInput, ImageInputProps, ImageValue } from "./ImageInput";
import React, { useState } from "react";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: ImageInput,
  tags: ["autodocs"],
  args: {
    value: null,
  },
  argTypes: {
    value: { table: { disable: true } },
    onValueChange: { table: { disable: true } },
  },
  render: (args) => {
    return <Demo {...args} />;
  },
} satisfies Meta<typeof ImageInput>;

function Demo({ value: _, ...args }: ImageInputProps) {
  const [value, setValue] = useState<ImageValue | null>(null);

  return (
    <div style={{ width: "600px" }}>
      <ImageInput value={value} onValueChange={setValue} {...args} />
    </div>
  );
}

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    sizeLimitInBytes: 100 * 1000,
  },
};
