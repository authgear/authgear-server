import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import type { IDropdownOption } from "@fluentui/react";
import FormDropdown from "./FormDropdown";

const options: IDropdownOption[] = [
  { key: "email", text: "Email" },
  { key: "phone", text: "Phone" },
  { key: "username", text: "Username" },
];

const meta = {
  title: "components/v1/FormDropdown",
  component: FormDropdown,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div style={{ width: 480 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    label: "Primary login identifier",
    parentJSONPointer: "",
    fieldName: "primary_login_id",
    options,
    selectedKey: "email",
    required: true,
  },
} satisfies Meta<typeof FormDropdown>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => {
    const [selectedKey, setSelectedKey] = useState(args.selectedKey);
    return (
      <FormDropdown
        {...args}
        selectedKey={selectedKey}
        onChange={(_, option) => setSelectedKey(option?.key)}
      />
    );
  },
};

export const WithPlaceholder: Story = {
  args: {
    selectedKey: undefined,
    placeholder: "Select one",
  },
  render: (args) => {
    const [selectedKey, setSelectedKey] = useState(args.selectedKey);
    return (
      <FormDropdown
        {...args}
        selectedKey={selectedKey}
        onChange={(_, option) => setSelectedKey(option?.key)}
      />
    );
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};
