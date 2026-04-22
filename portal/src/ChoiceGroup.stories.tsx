import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
// eslint-disable-next-line no-restricted-imports
import { ChoiceGroup, IChoiceGroupOption } from "@fluentui/react";

const sampleOptions: IChoiceGroupOption[] = [
  { key: "daily", text: "Daily digest" },
  { key: "weekly", text: "Weekly summary" },
  { key: "never", text: "No emails" },
];

const meta = {
  title: "components/v1/ChoiceGroup",
  component: ChoiceGroup,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div style={{ width: 440 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    label: "Notification frequency",
    options: sampleOptions,
  },
  argTypes: {
    onChange: { table: { disable: true } },
  },
} satisfies Meta<typeof ChoiceGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => {
    const [selectedKey, setSelectedKey] = useState<string | undefined>(
      "weekly"
    );
    return (
      <ChoiceGroup
        {...args}
        selectedKey={selectedKey}
        onChange={(_, option) => setSelectedKey(option?.key as string | undefined)}
      />
    );
  },
};

export const Horizontal: Story = {
  render: (args) => {
    const [selectedKey, setSelectedKey] = useState<string | undefined>(
      "card"
    );
    return (
      <ChoiceGroup
        {...args}
        selectedKey={selectedKey}
        onChange={(_, option) => setSelectedKey(option?.key as string | undefined)}
        styles={{
          flexContainer: {
            display: "flex",
            flexDirection: "row",
            flexWrap: "wrap",
            gap: "16px",
          },
        }}
      />
    );
  },
  args: {
    label: "Payment method",
    options: [
      { key: "card", text: "Card" },
      { key: "bank", text: "Bank transfer" },
      { key: "invoice", text: "Invoice" },
    ],
  },
};

export const WithDisabledOption: Story = {
  render: (args) => {
    const [selectedKey, setSelectedKey] = useState<string | undefined>(
      "standard"
    );
    return (
      <ChoiceGroup
        {...args}
        selectedKey={selectedKey}
        onChange={(_, option) => setSelectedKey(option?.key as string | undefined)}
      />
    );
  },
  args: {
    label: "Plan tier",
    options: [
      { key: "standard", text: "Standard" },
      { key: "pro", text: "Pro" },
      { key: "legacy", text: "Legacy (deprecated)", disabled: true },
    ],
  },
};

export const GroupDisabled: Story = {
  args: {
    label: "Unavailable (maintenance)",
    options: sampleOptions,
    defaultSelectedKey: "daily",
    disabled: true,
  },
};
