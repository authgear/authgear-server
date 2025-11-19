import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { RadioCards, RadioCardsProps } from "./RadioCards";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: RadioCards,
  tags: ["autodocs"],
  argTypes: {
    value: { table: { disable: true } },
    onValueChange: { table: { disable: true } },
  },
  args: {
    options: new Array(3).fill(null).map((_, idx) => ({
      value: `${idx + 1}`,
      title: `Title ${idx + 1}`,
    })),
    value: null,
    onValueChange: () => {},
  },
  render: (args) => {
    return <RadioCardsDemo args={args} />;
  },
} satisfies Meta<typeof RadioCards>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    size: "2",
    highContrast: false,
    options: new Array(3).fill(null).map((_, idx) => ({
      value: `${idx + 1}`,
      title: `Title ${idx + 1}`,
    })),
  },
};

export const DarkHighContrast: Story = {
  args: {
    darkMode: true,
    size: "2",
    highContrast: true,
    options: new Array(3).fill(null).map((_, idx) => ({
      value: `${idx + 1}`,
      title: `Title ${idx + 1}`,
    })),
  },
  parameters: {
    backgrounds: {
      default: "dark",
    },
  },
};

export const WithSubtitle: Story = {
  args: {
    size: "2",
    options: new Array(3).fill(null).map((_, idx) => ({
      value: `${idx + 1}`,
      title: `Title ${idx + 1}`,
      subtitle: `Subtitle ${idx + 1}`,
    })),
  },
};

export const Disabled: Story = {
  args: {
    size: "2",
    options: [
      {
        value: `1`,
        title: `Normal`,
        subtitle: `Subtitle`,
      },
      {
        value: `2`,
        title: `Disabled`,
        subtitle: `Subtitle`,
        disabled: true,
      },
    ],
  },
};

function RadioCardsDemo({ args }: { args: RadioCardsProps<string> }) {
  const { value: _0, onValueChange: _1, ...rest } = args;

  const [value, setValue] = useState<string | null>(null);

  return (
    <div style={{ width: 600 }}>
      <RadioCards value={value} onValueChange={setValue} {...rest} />
    </div>
  );
}
