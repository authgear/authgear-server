import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import Toggle from "./Toggle";

const meta = {
  title: "components/v1/Toggle",
  component: Toggle,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div style={{ width: 400 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    label: "Enable email notifications",
    inlineLabel: true,
  },
} satisfies Meta<typeof Toggle>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => {
    const { defaultChecked, ...rest } = args;
    const [checked, setChecked] = useState(defaultChecked ?? false);
    return (
      <Toggle
        {...rest}
        checked={checked}
        onChange={(_, v) => setChecked(v ?? false)}
      />
    );
  },
  args: {
    defaultChecked: true,
  },
};

export const WithDescription: Story = {
  render: (args) => {
    const { defaultChecked, ...rest } = args;
    const [checked, setChecked] = useState(defaultChecked ?? false);
    return (
      <Toggle
        {...rest}
        checked={checked}
        onChange={(_, v) => setChecked(v ?? false)}
      />
    );
  },
  args: {
    label: "Require verification",
    description:
      "When on, users must verify this channel before it can be used for sign-in.",
  },
};

export const Disabled: Story = {
  args: {
    label: "Locked setting",
    disabled: true,
    checked: false,
  },
};

export const DisabledOn: Story = {
  args: {
    label: "Locked setting (on)",
    disabled: true,
    checked: true,
  },
};

/**
 * Uses locale strings `Toggle.on` / `Toggle.off` beside the switch (portal default for dense rows).
 */
export const InlineLabelOffWithOnOffText: Story = {
  render: (args) => {
    const { defaultChecked, ...rest } = args;
    const [checked, setChecked] = useState(defaultChecked ?? true);
    return (
      <Toggle
        {...rest}
        checked={checked}
        onChange={(_, v) => setChecked(v ?? false)}
      />
    );
  },
  args: {
    label: "Passkey",
    inlineLabel: false,
  },
};
