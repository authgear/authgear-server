import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { TextField, TextFieldIcon } from "./TextField";

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
    hint: {
      control: {
        type: "text",
      },
    },
    iconStart: {
      options: ["none", ...Object.keys(TextFieldIcon)],
      mapping: {
        none: undefined,
        ...Object.keys(TextFieldIcon).reduce<Record<string, any>>(
          (mapping, it) => {
            mapping[it] = TextFieldIcon[
              it as keyof typeof TextFieldIcon
            ] as any;
            return mapping;
          },
          {}
        ),
      },
    },
    iconEnd: {
      options: ["none", ...Object.keys(TextFieldIcon)],
      mapping: {
        none: undefined,
        ...Object.keys(TextFieldIcon).reduce<Record<string, any>>(
          (mapping, it) => {
            mapping[it] = TextFieldIcon[
              it as keyof typeof TextFieldIcon
            ] as any;
            return mapping;
          },
          {}
        ),
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
    iconStart: TextFieldIcon.MagnifyingGlass,
    iconEnd: TextFieldIcon.InfoCircled,
  },
};

export const Suffix: Story = {
  args: {
    suffix: ".authgearapps.com",
  },
  render: (args) => {
    return (
      <div style={{ width: "560px" }}>
        <TextField {...args} />
      </div>
    );
  },
};

export const Hint: Story = {
  args: {
    hint: "This is some hint under the field.",
  },
  render: (args) => {
    return (
      <div style={{ width: "560px" }}>
        <TextField {...args} />
      </div>
    );
  },
};
