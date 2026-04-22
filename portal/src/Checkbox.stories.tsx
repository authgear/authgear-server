import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
// eslint-disable-next-line no-restricted-imports
import { Checkbox } from "@fluentui/react";

const meta = {
  title: "components/v1/Checkbox/Base",
  component: Checkbox,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div style={{ width: 400 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    label: "Remember this device",
  },
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => {
    const [checked, setChecked] = useState(false);
    return (
      <Checkbox
        {...args}
        checked={checked}
        onChange={(_, v) => setChecked(v ?? false)}
      />
    );
  },
};

export const Checked: Story = {
  render: (args) => {
    const [checked, setChecked] = useState(true);
    return (
      <Checkbox
        {...args}
        checked={checked}
        onChange={(_, v) => setChecked(v ?? false)}
      />
    );
  },
};

export const Disabled: Story = {
  args: {
    label: "Unavailable option",
    disabled: true,
    checked: false,
  },
};

export const DisabledChecked: Story = {
  args: {
    label: "Locked on",
    disabled: true,
    checked: true,
  },
};

export const Indeterminate: Story = {
  args: {
    label: "Select all (partial)",
    indeterminate: true,
    checked: false,
  },
};
