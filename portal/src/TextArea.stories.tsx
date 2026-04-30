import React, { ComponentProps, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import TextArea from "./TextArea";

const meta = {
  title: "components/v1/TextArea",
  component: TextArea,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div style={{ width: 480 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    label: "Notes",
    rows: 4,
    placeholder: "Optional context for reviewers…",
  },
} satisfies Meta<typeof TextArea>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithDescription: Story = {
  args: {
    label: "Allowed origins",
    description: "One origin per line. Must include scheme (https://).",
    rows: 6,
    placeholder: "https://app.example.com",
  },
};

function TextAreaControlledRender(args: ComponentProps<typeof TextArea>) {
  const [value, setValue] = useState(args.value ?? "");
  return (
    <TextArea {...args} value={value} onChange={(_, v) => setValue(v ?? "")} />
  );
}

export const Controlled: Story = {
  render: (args) => <TextAreaControlledRender {...args} />,
  args: {
    label: "Internal memo",
    value: "Initial content",
    rows: 4,
  },
};

export const ReadOnly: Story = {
  args: {
    label: "Generated policy",
    value: '{\n  "version": 1\n}',
    readOnly: true,
    rows: 5,
  },
};

export const Disabled: Story = {
  args: {
    label: "Disabled",
    value: "Cannot edit this block.",
    disabled: true,
    rows: 3,
  },
};

export const WithError: Story = {
  args: {
    label: "JSON configuration",
    value: "{ invalid",
    errorMessage: "Parse error: unexpected end of input.",
    rows: 4,
  },
};
